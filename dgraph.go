package mulbase

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gammazero/workerpool"
	"google.golang.org/grpc/encoding/gzip"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

//Deprecated information but useful.
//tlsPaths has length 5.
//First parameter is the root CA.
//Second is the client crt, third is the client key.
//Fourth is the node crt, fifth is the node key.

type Config struct {
	//For initialization.
	IP      string
	Port    int
	Tls     bool
	RootCA  string
	NodeCRT string
	NodeKey string
	//Other stuff.
	LogQueries bool
}

//DNode represents an object that can be safely stored
//in the database. It includes all necessary fields for
//automatic generation.
type DNode interface {
	//Returns the UID of this node.
	UID() UID
	//Sets the UID of this node.
	SetUID(uid UID)
	//Sets all types of this node. This has to be done at least once.
	SetType()
	//Returns all scalar fields for this node.
	Fields() FieldList
	//Serializes all the scalar values that are not hidden.
	Values() DNode
	//MapValues instead creates a map of values to allow you to build custom save methods.
	MapValues() Mapper
}
//Querier is an abstraction over DB/TXN. Also allows for testing.
type Querier interface {
	Query(context.Context, Query, ...interface{}) error
	Mutate(context.Context, Query) (*api.Response, error)
	Discard(context.Context) error
	Commit(context.Context) error
}

type AsyncQuerier interface {
	Querier
	QueryAsync(context.Context, Query, ...interface{}) chan Result
	MutateAsync(context.Context, Query)
}

func NewMapper(uid UID) Mapper {
	return Mapper{"uid": uid}
}

//UidObject is embedded in maps to ensure proper serialization.
type MapUid struct  {
	Uid UID `json:"uid"`
}
//Uid simply returns a struct that
//can be used in 1-1 relations, i.e.
//map[key] = mulbase.Uid(uiD)
func Uid(u UID) MapUid{
	return MapUid{Uid:u}
}

//A mapper that allows you to set subrelations.
type Mapper map[string]interface{}

func (m Mapper) UID() UID {
	return m["uid"].(UID)
}

func (m Mapper) SetUID(uid UID) {
	m["uid"] = UID(uid)
}

func (m Mapper) SetType() {
	fmt.Println("SetType called on Mapper. Is this intended?")
}

func (m Mapper) Fields() FieldList {
	return nil
}

func (m Mapper) Values() DNode {
	return m
}

func (m Mapper) MapValues() Mapper {
	return m
}

//Sets a singular regulation, i.e. 1-1.
//TODO: Should default be = obj or = obj.Values()?
func (m Mapper) Set(child Predicate, all bool, obj DNode) Mapper {
	if checkNil(obj) {
		//TODO: what should actually happen here?
		//panic for now.
		panic("nil not expected in Mapper set")
	}
	if all {
		if val, ok := obj.(Saver); ok {
			m[string(child)] = val.Save()
		} else {
			m[string(child)] = obj
		}
	} else {
		m[string(child)] = Uid(obj.UID())
	}
	return m
}
//honestly, Go nil is annoying for interfaces.
func checkNil(c DNode) bool {
	if c == nil || (reflect.ValueOf(c).Kind() == reflect.Ptr && reflect.ValueOf(c).IsNil()) {
		return true
	}
	return false
}

func (m Mapper) SetArray(child string, all bool, objs ...DNode) Mapper {
	var output = make([]interface{}, len(objs))
	for k,v := range objs {
		if all {
			if val, ok := v.(Saver); ok {
				output[k] = val.Save()
			} else {
				output[k] = val
			}
		} else {
			output[k] = v.UID()
		}
	}
	m[child] = output
	return m
}

//Saver allows you to implement a custom save method.
type Saver interface {
	Save() DNode
}

type Deleter interface {
	Delete() DNode
}

//Number of workers.
const workers = 5

type DB struct {
	//The api to graph.
	d *dgo.Dgraph
	//Config.
	c *Config
	//Schema list.
	schema SchemaList
	//The pool of asynchronous workers.
	pool *workerpool.WorkerPool
	//The endpoint for possible GraphQL.
	gplPoint string
}

func (d *DB) Schema() SchemaList {
	return d.schema
}

//NewTxn creates a new txn for interacting with database.
func (d *DB) NewTxn(readonly bool) *Txn {
	if d.schema == nil {
		panic("transaction without schema")
	}
	txn := new(Txn)
	if readonly {
		txn.txn = d.d.NewReadOnlyTxn()
	} else {
		txn.txn = d.d.NewTxn()
	}
	txn.db = d
	return txn
}

//Queries outside a Txn context.
//This is not intended for mutations.
func (d *DB) Query(ctx context.Context, q Query, objs ...interface{}) error {
	if q.Type() != QueryRegular {
		return Error(errInvalidType)
	}
	txn := d.NewTxn(true)
	err := txn.Query(ctx, q, objs...)
	//TODO: Can we do this for readonly?
	_ = txn.Discard(ctx)
	return err
}

func (d *DB) Commit(ctx context.Context) error {
	return nil
	//Moot
}

func (d *DB) Discard(ctx context.Context) error {
	return nil
	//Moot
}

func (d *DB) Mutate(ctx context.Context, q Query) (*api.Response, error) {
	if q.Type() != QuerySet && q.Type() != QueryDelete {
		return nil, Error(errInvalidType)
	}
	txn := d.NewTxn(false)
	resp, err := txn.Mutate(ctx, q)
	if err != nil {
		_ = txn.Discard(ctx)
	} else {
		_ = txn.Commit(ctx)
	}
	return resp, err
}

//Simply performs the alter command.
func (d *DB) Alter(ctx context.Context, op *api.Operation) error {
	err := d.d.Alter(ctx, op)
	return err
}

func (d *DB) logError(ctx context.Context, err error) {
	f := func() {
		errorName := fmt.Sprintf("%T", err)
		value := err.Error()
		var result dbError
		result.SetType()
		result.Message = value
		result.Time = time.Now()
		result.ErrorType = errorName
		_,_ = d.Mutate(context.Background(), CreateMutation(&result, QuerySet))
	}
	d.pool.Submit(f)
	return
}

type QueryType uint8

const (
	QueryRegular QueryType = iota
	QuerySet
	QueryDelete
)

//Allows it to be used in Query.
type Query interface {
	//Process the query type in order to send to the database.
	Process(SchemaList) ([]byte, map[string]string, error)
	//What type of query is this? Mutation(set/delete), regular query?
	Type() QueryType
}
//Init creates a new database connection using the given config
//and schema.
func Init(conf *Config, sch map[Predicate]Field) *DB {
	if conf.Port < 1000 {
		panic("graphinit: invalid dgraph port number")
	}
	db := connect(conf, sch)
	return db
}

//Takes a connection create of the form http://ip:port/
func connect(conf *Config, sch SchemaList) *DB {
	//TODO: allow multiple dgraph clusters.
	var conn *grpc.ClientConn
	var err error
	if conf.Tls {
		//TODO: Review this code as it might change how TLS works with dgraph.
		rootCAs := x509.NewCertPool()
		cCerts, err := tls.LoadX509KeyPair(conf.NodeCRT, conf.NodeKey)
		certs, err := ioutil.ReadFile(conf.RootCA)
		rootCAs.AppendCertsFromPEM(certs)
		tlsConf := &tls.Config{}
		tlsConf.RootCAs = rootCAs
		tlsConf.Certificates = append(tlsConf.Certificates, cCerts)
		c := credentials.NewTLS(tlsConf)
		conn, err = grpc.Dial(conf.IP+":"+strconv.Itoa(conf.Port), grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)), grpc.WithTransportCredentials(c))
		if err != nil {
			panic(err)
		}
	} else {
		conn, err = grpc.Dial(conf.IP+":"+strconv.Itoa(conf.Port), grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)), grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
	}
	//TODO: For multiple DGraph servers append multiple connections here.
	var c = dgo.NewDgraphClient(api.NewDgraphClient(conn))
	db := &DB{
		d:        c,
		gplPoint: conf.IP + ":" + strconv.Itoa(conf.Port) + "/graphql",
		c: conf,
		schema: sch,
	}
	//Run an empty query to ensure connection.
	db.pool = workerpool.New(workers)
	_ = db.Query(context.Background(), NewStaticQuery(""), nil)
	return db
}

//Txn is a non thread-safe API for interacting with the database.
//TODO: Should it be thread-safe?
//TODO: Do not keep a storage of previous queries and reuse GeneratedQuery with sync.Pool?
type Txn struct {
	//Allows for safe storage in Queries.
	sync.Mutex
	//All queries performed by this transaction.
	Queries []Query
	//The actual dgraph transaction.
	txn *dgo.Txn
	//The database. This is used for worker-pool & schema.
	db *DB
}

func (t *Txn) Commit(ctx context.Context) error {
	return t.txn.Commit(ctx)
}

func (t *Txn) Discard(ctx context.Context) error {
	return t.txn.Discard(ctx)
}

//Perform a single mutation.
func (t *Txn) mutate(ctx context.Context, q Query) (*api.Response, error) {
	//Add a single mutation to the query list.
	byt, _, err := q.Process(t.db.schema)
	if err != nil {
		return nil, err
	}
	var m api.Mutation
	if q.Type() == QueryDelete {
		//TODO: fix this
		m.DeleteJson = byt
	} else if q.Type() == QuerySet {
		m.SetJson = byt
	}
	a, err := t.txn.Mutate(ctx, &m)
	if err != nil {
		t.db.logError(context.Background(), err)
	}
	return a, err
}

//Upsert follows the new 1.1 api and performs an upsert.
//TODO: I really don't know what this does so work on it later.
func (t *Txn) Upsert(ctx context.Context, q Query, cond string, mutations ...Query) (int, error) {
	if t.txn == nil {
		return 0, Error(errTransaction)
	}
	b, ma, err := q.Process(t.db.schema)
	if err != nil {
		return 0, Error(err)
	}
	var muts = make([]*api.Mutation, len(mutations))
	for k := range muts {
		if mutations[k].Type() == QueryRegular {
			return 0, errInvalidType
		}
		muts[k] = new(api.Mutation)
		v := muts[k]
		b,_, err := mutations[k].Process(t.db.schema)
		if err != nil {
			return 0, err
		}
		v.SetJson = b
		v.Cond = cond
	}
	var req = api.Request{
		//TODO: Dont use create(b) as that performs unnecessary allocations. We do not perform any changes to b which should not cause any issues.
		Query:     bytesToStringUnsafe(b),
		Vars:      ma,
		Mutations: muts,
	}
	resp, err := t.txn.Do(ctx, &req)
	if err != nil {
		return 0, Error(err)
	}
	//err = HandleResponse(resp.Json, obj)
	return len(resp.Uids), nil
}

func (t *Txn) query(ctx context.Context, q Query, objs []interface{}) error {
	str, m, err := q.Process(t.db.schema)
	if err != nil {
		return err
	}
	resp, err := t.txn.QueryWithVars(ctx, bytesToStringUnsafe(str), m)
	if err != nil {
		t.db.logError(context.Background(), err)
		return Error(err)
	}
	if t.db.c.LogQueries {
		log.Printf("Query input: %s \n Query output: %s", string(str), string(resp.Json))
	}
	//This deserializes using reflect.
	err = HandleResponse(resp.Json, objs)
	if err != nil {
		return Error(err)
	}
	return nil
}

type Result struct {
	Err error
	Res *api.Response
}

//QueryAsync runs the query asynchronous. In the result the error is returned.
func (t *Txn) QueryAsync(ctx context.Context, q Query, objs ...interface{}) chan Result {
	ch := make(chan Result, 1)
	f := func() {
		err := t.Query(ctx, q, objs...)
		ch <- Result{
			Err: err,
		}
	}
	t.db.pool.Submit(f)
	return ch
}

//Query executes the GraphQL+- query.
//If q is a mutation query the mutation objects are supplied in q and not in objs.
func (t *Txn) Query(ctx context.Context, q Query, objs ...interface{}) error {
	//if t.txn == nil {
	//	return Error(errTransaction)
	//}
	//Allow thread-safe appending of queries as might run queries.
	//TODO: Right now this is only for storing the queries. Running queries in parallel that rely on each other is very much a race condition.
	t.Lock()
	defer t.Unlock()
	t.Queries = append(t.Queries, q)
	switch q.Type() {
	case QueryRegular:
		//No need for the response.
		return Error(t.query(ctx, q, objs))
	default:
		return errInvalidType
	}
}

func (t *Txn) Mutate(ctx context.Context, q Query) (*api.Response, error) {
	t.Lock()
	defer t.Unlock()
	typ := q.Type()
	if typ != QuerySet && typ != QueryDelete {
		t.Unlock()
		return nil, errInvalidType
	}
	t.Queries = append(t.Queries, q)
	return t.mutate(ctx, q)
}

func (t *Txn) MutateAsync(ctx context.Context, q Query) chan Result {
	typ := q.Type()
	if typ != QuerySet && typ != QueryDelete {
		return nil
	}
	var ch = make(chan Result)
	f := func() {
		resp, err := t.mutate(ctx, q)
		ch <- Result{Res:resp, Err:err}
	}
	t.db.pool.Submit(f)
	return ch
}
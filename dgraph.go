package humus

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"github.com/gammazero/workerpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Deprecated information but useful.
//tlsPaths has length 5.
//First parameter is the root CA.
//Second is the client crt, third is the client key.
//Fourth is the node crt, fifth is the node key.

//Config represents the configuration for connecting to a Dgraph alpha client.
type Config struct {
	//IP for this alpha.
	IP string
	//Port for this alpha.
	Port int
	//Tls connection.
	Tls bool
	//RootCA is the path for the RootCA.
	RootCA string
	//NodeCRT is the path for the NodeCRT.
	NodeCRT string
	//NodeKey is the path for the NodeKey.
	NodeKey string
	//Log queries in stdout along their result.
	LogQueries bool
}

//DNode represents an object that can be safely stored
//in the database. It includes all necessary fields for
//automatic generation.
//This is a big interface but it is automatically satisfied by all values.
type DNode interface {
	//Returns the UID of this node.
	UID() UID
	//Sets the UID of this node.
	SetUID(uid UID)
	//Sets all types of this node. This has to be done at least once.
	SetType()
	//Returns the type.
	GetType() []string
	//Returns all scalar fields for this node. This is not of fields interface
	//as they default nods always return a FieldList and not a NewList for instance.
	Fields() FieldList
	//Serializes all the scalar values that are not hidden. It usually returns
	//a type of *{{.Type}}Scalars.
	Values() DNode
	//Recurse allows you to set types and UIDS for all sub nodes.
	Recurse()
}

//Querier is an abstraction over DB/TXN. Also allows for testing.
type Querier interface {
	//Query queries the database with a variable amount of interfaces to deserialize into.
	//That is, if you are performing two queries q and q1 you are expected to supply two values.
	Query(context.Context, Query, ...interface{}) error
	//mutate mutates the query and returns the response.
	Mutate(context.Context, Query) (*api.Response, error)
	//Discard the transaction. This is done automatically in DB but not in Txn.
	Discard(context.Context) error
	//Commit is the same as above except it commits the transaction.
	Commit(context.Context) error
}

//AsyncQuerier is an interface representing a querier that can also
//perform queries asynchronously.
type AsyncQuerier interface {
	Querier
	QueryAsync(context.Context, Query, ...interface{}) chan Result
	MutateAsync(context.Context, Query) chan Result
}

//NewMapper returns a map as a Dnode. This is useful for building up
//custom structures. For completely custom mutations see CreateCustomMutation.
//Types are not mandatory but should be supplied for new nodes.
func NewMapper(uid UID, typ []string) Mapper {
	if len(typ) > 0 {
		return Mapper{"uid": uid, "dgraph.type": typ}
	}
	return Mapper{"uid": uid}
}

//Uid simply returns a struct that
//can be used in 1-1 relations, i.e.
//map[key] = mulbase.Uid(uiD)
func Uid(u UID) Node {
	return Node{Uid: u}
}

//Mapper allows you to set subrelations manually.
type Mapper map[string]interface{}

//UID returns the uid for this map.
func (m Mapper) UID() UID {
	if val, ok := m["uid"].(UID); ok {
		return val
	}
	return ""
}

type onlyUid struct {
	Uid UID `json:"uid"`
}

//SetFunctionValue is used in relation to upsert.
//For instance, storing uid in "result" variable. This will set the value
//on the edge predicate to the value. This is a simple case and for more complicated
//queries manually edit the Mapper to look correct.
//If predicate is "uid" do not set it as a child relation.
func (m Mapper) SetFunctionValue(variableName string, predicate Predicate) {
	uid := UID("uid(" + variableName + ")")
	if predicate == "uid" {
		m["uid"] = uid
	} else {
		m[string(predicate)] = onlyUid{uid}
	}
}

func (m Mapper) GetType() []string {
	fmt.Println("GetType called on Mapper. Is this intended?")
	return nil
}

func (m Mapper) SetUID(uid UID) {
	m["uid"] = uid
}

func (m Mapper) SetType() {
	fmt.Println("SetType called on Mapper. Is this intended?")
}

func (m Mapper) Fields() FieldList {
	return nil
}

func (m Mapper) Recurse() {

}

func (m Mapper) Values() DNode {
	return m
}

func (m Mapper) MapValues() Mapper {
	return m
}

//Set sets a singular regulation, i.e. 1-1.
//If all uses saver interface or the entire object. Otherwise, only uid is used.
func (m Mapper) Set(child Predicate, all bool, obj DNode) Mapper {
	if checkNil(obj) {
		//TODO: what should actually happen here?
		return m
	}
	obj.Recurse()
	if all {
		if val, ok := obj.(Saver); ok {
			m[string(child)] = val.Save()
		} else {
			m[string(child)] = obj
		}
	} else {
		//Default is only a uid edge. -> {"uid": "uid"}
		m[string(child)] = Uid(obj.UID())
	}
	return m
}

//MustSet panics on a nil node. Otherwise it is the same as Set.
func (m Mapper) MustSet(child Predicate, all bool, obj DNode) Mapper {
	if checkNil(obj) {
		//TODO: what should actually happen here?
		//panic for now.
		panic("mapper MustSet nil value")
	}
	obj.Recurse()
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

//SetArray sets a list of edges, saving as the form [node1, node2..].
func (m Mapper) SetArray(child string, all bool, objs ...DNode) Mapper {
	var output = make([]interface{}, len(objs))
	for k, v := range objs {
		v.Recurse()
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

//Deleter allows you to implement a custom delete method.
type Deleter interface {
	Delete() DNode
}

//Number of workers.
const workers = 5

//DB is the root object for using humus. It is used to immediately communicate with Dgraph
//as well as spawning new transactions. It handles a pool of dgraph clients as well as the active schema
//for the database.
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
	gplPoint      string
	interruptFunc func(i interface{})
}

//Schema returns the active schema for this database.
func (d *DB) Schema() SchemaList {
	return d.schema
}

//OnAborted sets the function for which to call
//when returned error is errAborted.
//The query variable is either a Mutate or a Query and should
//be handled accordingly.
func (d *DB) OnAborted(f func(query interface{})) {
	d.interruptFunc = f
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

//Select allows you to create a field list of only certain predicates.
//It is recommended to select from generated field lists as it s run at init
//and causes less overhead in queries.
func (d *DB) Select(vals ...Predicate) Fields {
	return Select(d.schema, vals...)
}

//Query queries outside a Txn context.
func (d *DB) Query(ctx context.Context, q Query, objs ...interface{}) error {
	txn := d.NewTxn(true)
	defer txn.Discard(context.Background())
	err := txn.Query(ctx, q, objs...)
	//TODO: Can we do this for readonly?
	if err != nil {
		fmt.Println(err)
	}
	return err
}

//QueryAsync runs the query and returns a channel with the result.
func (d *DB) QueryAsync(ctx context.Context, q Query, objs ...interface{}) chan Result {
	//No overhead so no point in deferring a discard in a function that does not make use of it.
	t := d.NewTxn(true)
	ch := make(chan Result, 1)
	f := func() {
		defer t.Discard(context.Background())
		err := t.Query(ctx, q, objs...)
		ch <- Result{
			Err: err,
		}
	}
	t.db.pool.Submit(f)
	return ch
}

//SetValue is a simple wrapper to set a single value in the database
//for a given node. It exists for convenience.
func (d *DB) SetValue(node DNode, pred Predicate, value interface{}) error {
	txn := d.NewTxn(true)
	defer txn.Discard(context.Background())
	var m = NewMapper(node.UID(), nil)
	m[string(pred)] = value
	_, err := txn.Mutate(context.Background(), CreateMutation(m, MutateSet))
	if err != nil {
		return err
	} else {
		return txn.Commit(context.Background())
	}
}

//Commit is a noop.
func (d *DB) Commit(ctx context.Context) error {
	return nil
	//Moot
}

//Discard is a noop.
func (d *DB) Discard(ctx context.Context) error {
	return nil
	//Moot
}

//Mutate runs a mutation outside a transaction context. It immediately commits on success.
func (d *DB) Mutate(ctx context.Context, m Mutate) (*api.Response, error) {
	txn := d.NewTxn(false)
	txn.commitNow = true
	defer txn.Discard(context.Background())
	resp, err := txn.Mutate(ctx, m)
	if err != nil {
		return nil, err
	}
	return resp, err
}

//Cleanup should be deferred at the main function.
func (d *DB) Cleanup() {
	d.pool.StopWait()
}

//Alter runs the command given by op to the dgraph instance.
func (d *DB) Alter(ctx context.Context, op *api.Operation) error {
	err := d.d.Alter(ctx, op)
	return err
}

func (d *DB) logError(ctx context.Context, err error) {
	f := func() {
		//Get the type name.
		if strings.Contains(err.Error(), "connection") {
			fmt.Println(err)
			return
		}
		errorName := fmt.Sprintf("%T", err)
		value := err.Error()
		var result dbError
		result.Recurse()
		result.Message = value
		result.Time = time.Now()
		result.ErrorType = errorName
		_, err := d.Mutate(context.Background(), CreateMutation(&result, MutateSet))
		if err != nil {
			fmt.Printf("Error logging %s \n", err.Error())
		}
	}
	d.pool.Submit(f)
	return
}

//Query is the humus interface for a query. GeneratedQuery, Queries, StaticQuery all satisfy this
//interface and is used in generating all necessary information
//to send the query to the database.
type Query interface {
	//Process the query type in order to send to the database.
	Process() (string, error)
	//What type of query is this? Mutation(set/delete), regular query?
	queryVars() map[string]string
	//names returns the names(or keys) for this query in order for deserialization.
	names() []string
}

//Mutate is the interface satisfied by all objects used in sending a json to the database.
//Any mutation sent to the database satisfies this interface.
type Mutate interface {
	mutate() ([]byte, error)
	Type() MutationType
	Cond() string
}

//Init is the entrypoint for humus. Init creates a a database object
//using the connection information as specified in the config.
//It uses the information specified to set up a grpc connection
//to the specified destination with or without TLS.
//It also sets the database schema using a pregenerated schema
//from humus/gen. Any query or mutation to the database goes through this object.
func Init(conf *Config, sch SchemaList) *DB {
	if conf.Port < 1000 {
		panic("graphinit: invalid dgraph port number")
	}
	db := connect(conf, sch)
	return db
}

func connect(conf *Config, sch SchemaList) *DB {
	//TODO: allow multiple dgraph clusters.
	var conn *grpc.ClientConn
	var err error
	if conf.Tls {
		//TODO: Review this code as it has not been tried in some time.
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
		c:        conf,
		schema:   sch,
	}
	//Run an empty query to ensure connection.
	db.pool = workerpool.New(workers)
	//_ = db.Query(context.Background(), NewStaticQuery(""), nil)
	return db
}

//Txn is a non thread-safe API for interacting with the database.
//TODO: Should it be thread-safe?
//TODO: Do not keep a storage of previous queries and reuse GeneratedQuery with sync.Pool?
type Txn struct {
	//Allows for safe storage in Queries.
	sync.Mutex
	//All queries performed by this transaction.
	Queries []interface{}
	//The actual dgraph transaction.
	txn *dgo.Txn
	//The database. This is used for worker-pool & schema.
	db *DB
	//To immediately commit mutations inside this txn.
	commitNow bool
}

func (t *Txn) Commit(ctx context.Context) error {
	return t.txn.Commit(ctx)
}

func (t *Txn) Discard(ctx context.Context) error {
	return t.txn.Discard(ctx)
}

//Perform a single mutation.
func (t *Txn) mutate(ctx context.Context, q Mutate) (*api.Response, error) {
	//Add a single mutation to the query list.
	byt, err := q.mutate()
	if err != nil {
		return nil, err
	}
	var m api.Mutation
	typ := q.Type()
	if typ == MutateSet {
		m.SetJson = byt
		if t.db.c.LogQueries {
			fmt.Println(string(m.SetJson))
		}
	} else if typ == MutateDelete {
		m.DeleteJson = byt
		if t.db.c.LogQueries {
			fmt.Println(string(m.DeleteJson))
		}
	}
	m.CommitNow = t.commitNow
	a, err := t.txn.Mutate(ctx, &m)
	if err != nil {
		t.db.logError(context.Background(), err)
	}
	if err == dgo.ErrAborted && t.db.interruptFunc != nil {
		t.db.interruptFunc(q)
	}
	return a, err
}

//Upsert follows the new 1.1 api and performs an upsert.
//q is a nameless query. For now it is recommended to use a static query for all upserts.
//Cond is a condition of the form if (eq(len(a), 0) and so on. mutations is a list of mutations to perform.
func (t *Txn) Upsert(ctx context.Context, q Query, mutations ...Mutate) (*api.Response, error) {
	if t.txn == nil {
		return nil, Error(errTransaction)
	}
	b, err := q.Process()
	if err != nil {
		return nil, Error(err)
	}
	var muts = make([]*api.Mutation, len(mutations))
	for k := range muts {
		muts[k] = new(api.Mutation)
		v := muts[k]
		b, err := mutations[k].mutate()
		if err != nil {
			return nil, err
		}
		if mutations[k].Type() == MutateDelete {
			v.DeleteJson = b
		} else {
			v.SetJson = b
		}
		v.Cond = mutations[k].Cond()
	}
	var req = api.Request{
		Query:     b,
		Vars:      q.queryVars(),
		Mutations: muts,
	}
	resp, err := t.txn.Do(ctx, &req)
	if err != nil {
		return nil, Error(err)
	}
	//err = HandleResponse(resp.Json, obj)
	return resp, nil
}

func (t *Txn) query(ctx context.Context, q Query, objs []interface{}) error {
	str, err := q.Process()
	if err != nil {
		return err
	}
	names := q.names()
	if len(names) != len(objs) {
		return Error(errors.New("mismatched length between query amount and input interfaces"))
	}
	if t.db.c.LogQueries {
		log.Printf("Query input: %s \n", str)
	}
	resp, err := t.txn.QueryWithVars(ctx, str, q.queryVars())
	if err == dgo.ErrAborted && t.db.interruptFunc != nil {
		t.db.interruptFunc(q)
	}
	if err != nil {
		//t.db.logError(context.Background(), err)
		return Error(err)
	}
	if t.db.c.LogQueries {
		log.Printf("Query output: %s", string(resp.Json))
	}
	//This deserializes using reflect.
	err = handleResponse(resp.Json, objs, names)
	//TODO: Ignore this error for now. Some oddity in dgraph when time is defaulted.
	if _, ok := err.(*time.ParseError); ok {
		return nil
	}
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
	t.Lock()
	defer t.Unlock()
	t.Queries = append(t.Queries, q)
	return Error(t.query(ctx, q, objs))
}

func (t *Txn) Mutate(ctx context.Context, q Mutate) (*api.Response, error) {
	t.Lock()
	defer t.Unlock()
	t.Queries = append(t.Queries, q)
	return t.mutate(ctx, q)
}

func (t *Txn) MutateAsync(ctx context.Context, q Mutate) chan Result {
	var ch = make(chan Result)
	f := func() {
		resp, err := t.mutate(ctx, q)
		ch <- Result{Res: resp, Err: err}
	}
	t.db.pool.Submit(f)
	return ch
}

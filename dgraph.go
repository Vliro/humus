package mulbase

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/gammazero/workerpool"
	"google.golang.org/grpc/encoding/gzip"
	"io/ioutil"
	"log"
	"strconv"
	"sync"

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
	SetUID(uid string)
	//Sets all types of this node. This has to be done at least once.
	SetType()
	//Returns all scalar fields for this node.
	Fields() FieldList
	//Serializes all the scalar values that are not hidden.
	Values() DNode
}

//Number of workers.
const workers = 10

type DB struct {
	//The api to graph.
	d *dgo.Dgraph
	//Config.
	c *Config
	//Schema list.
	schema schemaList
	//The pool of asynchronous workers.
	pool *workerpool.WorkerPool
	//The endpoint for possible GraphQL.
	gplPoint string
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
func (d *DB) Query(ctx context.Context, q Query, obj interface{}) error {
	if q.Type() != QueryRegular {
		return Error(errInvalidType)
	}
	txn := d.NewTxn(true)
	_, err := txn.RunQuery(ctx, q, obj)
	//TODO: Can we do this for readonly?
	_ = txn.Discard(ctx)
	return err
}

func (d *DB) Mutate(ctx context.Context, q Query) error {
	if q.Type() != QuerySet || q.Type() != QueryDelete {
		return Error(errInvalidType)
	}
	txn := d.NewTxn(false)
	_, err := txn.RunQuery(ctx, q, nil)
	if err != nil {
		_ = txn.Discard(ctx)
	} else {
		_ = txn.Commit(ctx)
	}
	return err
}

//Simply performs the alter command.
func (d *DB) Alter(ctx context.Context, op *api.Operation) error {
	err := d.d.Alter(ctx, op)
	return err
}

type QueryType uint8

const (
	QueryRegular QueryType = iota
	QuerySet
	QueryDelete
)

//Allows it to be used in runQuery.
type Query interface {
	//Process the query type in order to send to the database.
	Process(schemaList) ([]byte, map[string]string, error)
	//What type of query is this? Mutation(set/delete), regular query?
	Type() QueryType
}
//Init creates a new database connection using the given config
//and schema.
func Init(conf *Config, sch map[string]Field) *DB {
	if conf.Port < 1000 {
		panic("graphinit: invalid dgraph port number")
	}
	db := connect(conf, sch)
	db.pool = workerpool.New(workers)
	return db
}

//Takes a connection create of the form http://ip:port/
func connect(conf *Config, sch schemaList) *DB {
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
	db.Query(context.Background(), NewStaticQuery(""), nil)
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
	return a, err
}

//Upsert follows the new 1.1 api and performs an upsert.
//TODO: I really don't know what this does so work on it later.
func (t *Txn) Upsert(ctx context.Context, q Query, m []*api.Mutation, obj ...interface{}) error {
	if t.txn == nil {
		return Error(errTransaction)
	}
	b, ma, err := q.Process(t.db.schema)
	if err != nil {
		return Error(err)
	}
	var req = api.Request{
		//TODO: Dont use create(b) as that performs unnecessary allocations. We do not perform any changes to b which should not cause any issues.
		Query:     bytesToStringUnsafe(b),
		Vars:      ma,
		Mutations: m,
	}
	resp, err := t.txn.Do(ctx, &req)
	if err != nil {
		return Error(err)
	}
	err = HandleResponse(resp.Json, obj)
	return Error(err)
}

func (t *Txn) query(ctx context.Context, q Query, objs []interface{}) error {
	str, m, err := q.Process(t.db.schema)
	if err != nil {
		return err
	}
	resp, err := t.txn.QueryWithVars(ctx, bytesToStringUnsafe(str), m)
	if err != nil {
		return Error(err)
	}
	if t.db.c.LogQueries {
		log.Printf("Query input: %s \n\n Query output: %s", string(str), string(resp.Json))
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

//RunQueryAsync runs the query asynchronous. In the result the error is returned.
func (t *Txn) RunQueryAsync(ctx context.Context, q Query, objs ...interface{}) chan Result {
	ch := make(chan Result, 1)
	f := func() {
		r, err := t.RunQuery(ctx, q, objs...)
		ch <- Result{
			Err: err,
			Res: r,
		}
	}
	t.db.pool.Submit(f)
	return ch
}

//RunQuery executes the GraphQL+- query.
//If q is a mutation query the mutation objects are supplied in q and not in objs.
func (t *Txn) RunQuery(ctx context.Context, q Query, objs ...interface{}) (*api.Response, error) {
	//if t.txn == nil {
	//	return Error(errTransaction)
	//}
	//Allow thread-safe appending of queries as might run queries.
	//TODO: Right now this is only for storing the queries. Running queries in parallel that rely on each other is very much a race condition.
	t.Lock()
	t.Queries = append(t.Queries, q)
	t.Unlock()
	switch q.Type() {
	case QueryRegular:
		//No need for the response.
		return nil, Error(t.query(ctx, q, objs))
	case QuerySet, QueryDelete:
		r, err := t.mutate(ctx, q)
		return r, Error(err)
	}
	return nil, Error(errInvalidType)
}

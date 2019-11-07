package mulbase

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc/encoding/gzip"
	"io/ioutil"
	"strconv"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type DB struct {
	d *dgo.Dgraph
	ip string
	port int
	tls bool
	endPoint string
}
//NewTxn creates a new txn for interacting with database.
func (d *DB) NewTxn(readonly bool) *Txn {
	txn := new(Txn)
	if readonly {
		txn.txn = d.d.NewReadOnlyTxn()
	} else {
		txn.txn = d.d.NewTxn()
	}
	return txn
}
//Queries outside a Txn context.
//This is not intended for mutations.
func (d *DB) Query(ctx context.Context, q Query, obj interface{}) error {
	if q.Type() != QueryRegular {
		return Error(errInvalidType)
	}
	txn := d.NewTxn(true)
	err := txn.RunQuery(ctx, q, obj)
	//TODO: Can we do this for readonly?
	_ = txn.Discard(ctx)
	return err
}

func (d *DB) Mutate(ctx context.Context, q Query) error {
	if q.Type() != QuerySet || q.Type() != QueryDelete {
		return Error(errInvalidType)
	}
	txn := d.NewTxn(false)
	err := txn.RunQuery(ctx, q, nil)
	if err != nil {
		_ = txn.Discard(ctx)
	} else {
		_ = txn.Commit(ctx)
	}
	return err
}

type QueryType uint8

const (
	QueryRegular QueryType = iota
	QuerySet
	QueryDelete
)

type Query interface {
	Process() ([]byte, map[string]string, error)
	Type() QueryType
}

//tlsPaths has length 5.
//First parameter is the root CA.
//Second is the client crt, third is the client key.
//Fourth is the node crt, fifth is the node key.
func Init(dip string, dport int, tls bool, tlsPath ...string) {
	if dport < 1000 {
		panic("graphinit: invalid dgraph port number")
	}
	connect(dip, dport, tls, tlsPath)
	initSchema()
}
//Takes a connection string of the form http://ip:port/
func connect(ip string, port int, dotls bool, paths []string) *DB{
	//TODO: allow multiple dgraph clusters.
	var conn *grpc.ClientConn
	var err error
	if dotls {
		//TODO: Review this code as it might change how TLS works with dgraph.
		rootCAs := x509.NewCertPool()
		cCerts, err := tls.LoadX509KeyPair(paths[3], paths[4])
		certs, err := ioutil.ReadFile(paths[0])
		rootCAs.AppendCertsFromPEM(certs)
		conf := &tls.Config{}
		conf.RootCAs = rootCAs
		conf.Certificates = append(conf.Certificates, cCerts)
		c := credentials.NewTLS(conf)
		conn, err = grpc.Dial(ip+":"+strconv.Itoa(port), grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)), grpc.WithTransportCredentials(c))
		if err != nil {
			panic(err)
		}
	} else {
		conn, err = grpc.Dial(ip+":"+strconv.Itoa(port), grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)), grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
	}
	//TODO: For multiple DGraph servers append multiple connections here.
	var c = dgo.NewDgraphClient(api.NewDgraphClient(conn))
	db := &DB{
		d:        c,
		ip:       ip,
		port:     port,
		tls:      dotls,
		endPoint: ip + ":" + strconv.Itoa(port) + "/graphql",
	}
	return db
}
//Txn is a non thread-safe API for interacting with the database.
//TODO: Should it be thread-safe?
type Txn struct {
	//All queries performed by this transaction.
	Queries []Query
	//The actual dgraph transaction.
	txn     *dgo.Txn
	counter int
}

func (t *Txn) Commit(ctx context.Context) error {
	return t.txn.Commit(ctx)
}

func (t *Txn) Discard(ctx context.Context) error {
	return t.txn.Discard(ctx)
}
//Perform a single mutation.
func (t *Txn) mutate(ctx context.Context, q Query) error {
	//Add a single mutation to the query list.
	byt,_, err := q.Process()
	if err != nil {
		return err
	}
	var m api.Mutation
	if q.Type() == QueryDelete {
		//TODO: fix this
		m.DeleteJson = byt
	} else if q.Type() == QuerySet{
		m.SetJson = byt
	}
	_, err = t.txn.Mutate(ctx, &m)
	return err
}
//Upsert follows the new 1.1 api and performs an upsert.
//TODO: I really don't know what this does so work on it later.
func (t *Txn) Upsert(ctx context.Context, q Query, m []*api.Mutation, obj ...interface{}) error {
	if t.txn == nil {
		return Error(errTransaction)
	}
	b, ma, err := q.Process()
	if err != nil {
		return Error(err)
	}
	var req = api.Request{
		//TODO: Dont use string(b) as that performs unnecessary allocations. We do not perform any changes to b.
		Query:                bytesToStringUnsafe(b),
		Vars:                 ma,
		Mutations:            m,
	}
	resp, err := t.txn.Do(ctx, &req)
	if err != nil {
		return Error(err)
	}
	err = HandleResponse(resp.Json, obj)
	return Error(err)
}

func (t *Txn) query(ctx context.Context, q Query, objs []interface{}) error{
	str, m, err := q.Process()
	if err != nil {
		return err
	}
	resp, err := t.txn.QueryWithVars(ctx,bytesToStringUnsafe(str),m)
	if err != nil {
		return Error(err)
	}
	err = HandleResponse(resp.Json, objs)
	if err != nil {
		return Error(err)
	}
	return nil
}

//RunQuery executes the GraphQL+- query.
//If q is a mutation query I expect objs to be supplied in q.
func (t *Txn) RunQuery(ctx context.Context, q Query, objs ...interface{}) error {
	if t.txn == nil {
		return Error(errTransaction)
	}
	t.Queries = append(t.Queries, q)
	t.counter++
	switch q.Type() {
	case QueryRegular:
		return Error(t.query(ctx, q, objs))
	case QueryDelete, QuerySet:
		return Error(t.mutate(ctx, q))
	}
	return Error(errInvalidType)
	/*m := make(map[string]string)
	//TODO: multiple GraphQL variables.
	for k, v := range q.VarMap {
		switch g := v.val.(type) {
		case string:
			m[k] = g
			break
		case int:
			m[k] = strconv.Itoa(g)
			break
		}
	}*/
}
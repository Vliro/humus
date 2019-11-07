package ngraph

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"

	"google.golang.org/grpc/encoding/gzip"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	//The number of routines that send mutations to the database.
	WorkerRoutines int = 5
)

var (
	//The dgraph connection pool
	c *dgo.Dgraph
	//Query buffer for background mutations.
	buf chan *Query
	//Mutex to protect failed mutations.
	mutex       sync.Mutex
	queryBuffer []*Query

	//Dgraph connection information.
	ip       string
	port     int
	dotls    bool
	tlsPaths []string
)

//Start the mutate routines.
func init() {
	buf = make(chan *Query)
	for k := 0; k < WorkerRoutines; k++ {
		go StartBufferLoop()
	}
}

//tlsPaths has length 5.
//First parameter is the root CA.
//Second is the client crt, third is the client key.
//Fourth is the node crt, fifth is the node key.
func GraphInit(dip string, dport int, tls bool, tlsPath ...string) {
	if dport < 1000 {
		panic("graphinit: invalid dgraph port number")
	}
	ip = dip
	port = dport
	dotls = tls
	tlsPaths = tlsPath
	connect()
	initSchema()
	go StartBufferLoop()
}

//Synchronous saves to database.
func StartBufferLoop() {
	for {
		q := <-buf
		_, err := q.Execute(nil, nil)
		if err != nil {
			appendQuery(q)
		}
	}
}

func getClient() *dgo.Dgraph {
	if c == nil {
		connect()
	}
	return c
}

func PerformQueries(q *Queries, inp []interface{}) error {
	txn := getClient().NewReadOnlyTxn()
	ctx := context.Background()
	defer txn.Discard(ctx)
	str, err := q.Stringify()
	if err != nil {
		return err
	}
	m := make(map[string]string)

	for _, v := range q.Queries {
		for i, iv := range v.VarMap {
			m[i] = iv.val.(string)
		}
	}

	resp, err := txn.QueryWithVars(ctx, str, m)
	if err != nil {
		return err
	}
	err = HandleResponseArray(resp.Json, inp)
	return err
}

func executeFix(inp []interface{}, str string, m map[string]string) error {
	txn := getClient().NewReadOnlyTxn()
	ctx := context.Background()
	defer txn.Discard(ctx)

	resp, err := txn.QueryWithVars(ctx, str, m)
	if err != nil {
		return err
	}
	return HandleResponseArray(resp.Json, inp)
}

//runQuery runs the given query agains the database.
func (q *Query) runQuery(ctx context.Context, inp interface{}) (map[string]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var txn *dgo.Txn
	if q.activeTxn != nil {
		txn = q.activeTxn
	} else {
		switch q.Type {
		case TypeMutate:
			txn = getClient().NewTxn()
			break
		case TypeQuery:
			txn = getClient().NewReadOnlyTxn()
		default:
			panic("query: invalid type")
		}
		defer txn.Discard(ctx)
	}
	//If we have a json mutation force save it instead.
	//JSONMutation is a special mutation wherein the entire mutate object is just saved with pre-serializing.
	switch q.Type {
	case TypeMutate:
		if len(q.Mutations) > 0 {
			m := api.Mutation{}
			js, err := q.JSON(false)
			if err != nil {
				return nil, err
			}
			if len(js) > 1 {
				if q.getMutateType == MutateSet {
					m.SetJson = js
				} else {
					m.DeleteJson = js
				}
				if q.activeTxn == nil {
					m.CommitNow = true
				}
				c2, err := txn.Mutate(ctx, &m)
				if err != nil {
					fmt.Printf("could not mutate, %s", err.Error())
					appendQuery(q)
					return nil, err
				}
				return c2.Uids, nil
			}
		}
		return nil, fmt.Errorf("empty mutation")
	case TypeQuery:
		str, err := q.JSON(true)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(str))
		m := make(map[string]string)
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
		}
		resp, err := txn.QueryWithVars(ctx, bytesToStringUnsafe(str), m)
		if err != nil {
			fmt.Println("error dgraph: ", err)
			return nil, err
		}
		if q.Deserialize {
			err = HandleResponse(resp.Json, inp)
		} else {
			GetResponse(resp.Json, inp)
		}
		return nil, err
	}
	return nil, errors.New("invalid operation type")
}

func appendQuery(q *Query) {
	mutex.Lock()
	queryBuffer = append(queryBuffer, q)
	mutex.Unlock()
}

//Takes a connection string of the form http://ip:port/
func connect() {
	if c != nil {
		return
	}
	//TODO: allow multiple dgraph clusters.
	var conn *grpc.ClientConn
	var err error
	if dotls {
		//TODO: Review this code as it might change how TLS works with dgraph.
		rootCAs := x509.NewCertPool()
		cCerts, err := tls.LoadX509KeyPair(tlsPaths[3], tlsPaths[4])
		certs, err := ioutil.ReadFile(tlsPaths[0])
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
	c = dgo.NewDgraphClient(api.NewDgraphClient(conn))
}

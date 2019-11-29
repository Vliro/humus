package mulbase

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/protos/api"
	"testing"
)

type testQuerier struct {
	expected string
}

func (t testQuerier) Query(c context.Context, q Query, vals ...interface{}) error {
	qu, err := q.Process(SchemaList{})
	if qu != t.expected {
		fmt.Println(qu)
		return errors.New("query failed")
	}
	return err
}

func (t testQuerier) Mutate(context.Context, Query) (*api.Response, error) {
	panic("implement me")
}

func (t testQuerier) Discard(context.Context) error {
	return nil
}

func (t testQuerier) Commit(context.Context) error {
	return nil
}

func newTest(expected string) testQuerier {
	return testQuerier{expected:expected}
}

var fields = ErrorFields.Sub(ErrorMessageField, ErrorFields.Sub(ErrorMessageField, ErrorFields))

func TestQuery(t *testing.T) {
	const expected = "query t($0:string,$1:string){q(func: eq(<Error.message>,$0)){Error.message (first:1,after:1)(orderasc: Error.time))@filter(eq(<Error.message>,$1))@facets{Error.message (first:1){Error.message  Error.errorType  Error.time uid} Error.errorType  Error.time uid} Error.errorType (first:1) Error.time (first:1) uid}}"
	q := NewQuery().SetFields(fields)
	q.SetFunction(MakeFunction(FunctionEquals).AddPredValue(ErrorMessageField, "Test"))
	q.AddSubCount(CountFirst, ErrorMessageField, 1)
	q.AddSubCount(CountFirst, ErrorTimeField, 1)
	q.AddSubCount(CountFirst, ErrorErrorTypeField, 1)
	q.AddSubCount(CountAfter, ErrorMessageField, 1)
	q.AddOrdering(OrderAsc, ErrorMessageField, ErrorTimeField)
	q.AddSubCount(CountFirst, ErrorMessageField + ErrorMessageField, 1)
	q.AddAggregation(TypeSum, ErrorMessageField, "test")
	q.AddSubFilter(MakeFunction(FunctionEquals).AddPredValue(ErrorMessageField, "Test"), ErrorMessageField)
	q.Facets(ErrorMessageField)
	//q.SetLanguage(LanguageDefault, true)

	err := newTest(expected).Query(context.Background(), q)
	if err != nil {
		t.Fail()
	}
}

func BenchmarkQuery(b *testing.B) {
	//f, err := os.Create("pprof")
	//if err != nil {
	//	panic(err)
	//}
	//pprof.StartCPUProfile(f)
	for i := 0; i < b.N; i++ {
		q := NewQuery().SetFields(fields)
		q.SetFunction(MakeFunction(FunctionEquals).AddPredValue(ErrorMessageField, "Test"))
		q.AddSubCount(CountFirst, ErrorMessageField, 1)
		q.AddSubCount(CountFirst, ErrorTimeField, 1)
		q.AddSubCount(CountFirst, ErrorErrorTypeField, 1)
		q.AddSubCount(CountAfter, ErrorMessageField, 1)
		q.AddSubCount(CountFirst, ErrorMessageField + ErrorMessageField, 1)
		q.AddAggregation(TypeSum, ErrorMessageField, "test")
		q.Facets(ErrorMessageField)
		_, _ = q.Process(SchemaList{})
	}
	b.ReportAllocs()
	//pprof.StopCPUProfile()
}
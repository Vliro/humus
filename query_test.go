package humus

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/protos/api"
	"testing"
)

/*
As of right now the tests are performed against static strings.
Eventually the tests shall run against a dgraph instance with data provided.
This will be done before it is released.
*/

type testQuerier struct {
	expected string
}

func (t testQuerier) Query(c context.Context, q Query, vals ...interface{}) error {
	qu, err := q.Process()
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
	return testQuerier{expected: expected}
}

func TestBigQuery(t *testing.T) {
	q := NewQuery(fields)
	q.Function(Equals).PredValue(ErrorMessageField, "Test")
	q.Count(CountFirst, ErrorMessageField, 1)
	q.Count(CountFirst, ErrorTimeField, 1)
	q.Count(CountFirst, ErrorErrorTypeField, 1)
	q.Count(CountAfter, ErrorMessageField, 1)
	q.Order(Ascending, ErrorMessageField, ErrorTimeField)
	q.Count(CountFirst, ErrorMessageField+ErrorMessageField, 1)
	q.Agg(Sum, ErrorMessageField, "test", "")
	q.Directive(Cascade)
	q.Directive(Normalize)
	q.Filter(MakeFilter(Equals).PredValue(ErrorMessageField, "Test"), ErrorMessageField)
	q.Filter(MakeFilter(Equals).PredValue(ErrorMessageField, "Test"), ErrorErrorTypeField)
	q.Facets(ErrorMessageField)
	str, _ := q.Process()
	fmt.Println(str)
}

var fields = ErrorFields.Sub(ErrorMessageField, ErrorFields)

func TestQuery(t *testing.T) {
	const expected = "query t($0:string,$1:string,$2:string){q(func: eq(<Error.message>,$2),orderasc: Error.time)@cascade@normalize{Error.message@filter(eq(<Error.message>,$0))(first:1,after:1)(orderasc: Error.time)@facets{Error.message(first:1) Error.errorType Error.time varr as ErrorTimeField  test : sum(val(varr)) uid} Error.errorType@filter(eq(<Error.message>,$1))(first:1) Error.time(first:1)  uid}}"
	q := NewQuery(fields)
	q.Function(Equals).PredValue(ErrorMessageField, "Test")
	q.Count(CountFirst, ErrorMessageField, 1)
	q.Count(CountFirst, ErrorTimeField, 1)
	q.Count(CountFirst, ErrorErrorTypeField, 1)
	q.Count(CountAfter, ErrorMessageField, 1)
	q.Order(Ascending, ErrorMessageField, ErrorTimeField)
	q.Count(CountFirst, ErrorMessageField+ErrorMessageField, 1)
	q.Agg(Sum, ErrorMessageField, "varr", "test")
	q.Variable("varr", ErrorMessageField, "ErrorTimeField", false)
	q.Order(Ascending, "", ErrorTimeField)
	q.Directive(Cascade)
	q.Directive(Normalize)
	q.Filter(MakeFilter(Equals).PredValue(ErrorMessageField, "Test"), ErrorMessageField)
	q.Filter(MakeFilter(Equals).PredValue(ErrorMessageField, "Test"), ErrorErrorTypeField)
	q.Facets(ErrorMessageField)
	q.Language(LanguageEnglish, true)
	err := newTest(expected).Query(context.Background(), q)
	if err != nil {
		t.Fail()
	}
}

func BenchmarkQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := NewQuery(fields)
		q.Function(Equals).PredValue(ErrorMessageField, "Test")
		q.Count(CountFirst, ErrorMessageField, 1)
		q.Count(CountFirst, ErrorTimeField, 1)
		q.Count(CountFirst, ErrorErrorTypeField, 1)
		q.Count(CountAfter, ErrorMessageField, 1)
		q.Order(Ascending, ErrorMessageField, ErrorTimeField)
		q.Count(CountFirst, ErrorMessageField+ErrorMessageField, 1)
		q.Agg(Sum, ErrorMessageField, "varr", "swag")
		q.Directive(Cascade)
		q.Directive(Normalize)
		q.Filter(MakeFilter(Equals).PredValue(ErrorMessageField, "Test"), ErrorMessageField)
		q.Filter(MakeFilter(Equals).PredValue(ErrorMessageField, "Test"), ErrorErrorTypeField)
		q.Facets(ErrorMessageField)
		_, _ = q.Process()
	}
	b.ReportAllocs()
}

const staticString = `query {
	a as var(func: uid(%s))
`

func BenchmarkStaticQuery(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var stat = NewStaticQuery(fmt.Sprintf(staticString, "0x2"))
		_, _ = stat.Process()
	}
	b.ReportAllocs()
}

func TestQueryFilter(t *testing.T) {
	//expected := "query t($0:string,$1:string){q(func: eq(<Error.time>,$0))@filter(eq(<Error.time>,$1)){Error.message Error.errorType Error.time uid}}"
	q := NewQuery(ErrorFields)
	q.Function(Equals).PredValue(ErrorTimeField, "testFunction")
	q.Filter(MakeFilter(Equals).PredValue(ErrorTimeField, "testFilter"), "")
	q.Order(Ascending, "", ErrorTimeField)
	q.Count(CountFirst, "", 5)
	str, err := q.Process()
	if err != nil {
		t.Fatal()
	}
	fmt.Println(str)
}

func TestQueries(t *testing.T) {
	const expected = "query t($0:int,$2:string,$1:string,$3:string){q1(func: eq(<Error.time>,$2),first:5,orderasc: Error.time){Error.message@filter(eq(<Error.time>,$0)) Error.errorType Error.time  uid}}{q2(func: eq(<Error.time>,$3),first:5,orderasc: Error.time){Error.message Error.errorType Error.time@filter(eq(<Error.time>,$1))  uid}}"
	list := NewQueries()
	q := list.NewQuery(ErrorFields)
	q.Function(Equals).PredValue(ErrorTimeField, "testFunctionOne")
	q.Filter(MakeFilter(Equals).PredValue(ErrorTimeField, 5), ErrorMessageField)
	q.Order(Ascending, "", ErrorTimeField)
	q.Count(CountFirst, "", 5)
	qu := list.NewQuery(ErrorFields)
	qu.Function(Equals).PredValue(ErrorTimeField, "testFunctionTwo")
	qu.Filter(MakeFilter(Equals).PredValue(ErrorTimeField, "filterOne"), ErrorTimeField)
	qu.Order(Ascending, "", ErrorTimeField)
	qu.Count(CountFirst, "", 5)
	str, err := list.Process()
	if err != nil {
		t.Fail()
		return
	}
	if str != expected {
		t.Fail()
		return
	}
	//Check variables to ensure the variable mapping is not broken.
	one := q.function.Variables[1]
	if list.vars[one.Value] != "testFunctionOne" {
		t.Fail()
		return
	}
	two := qu.function.Variables[1]
	if list.vars[two.Value] != "testFunctionTwo" {
		t.Fail()
		return
	}
	three := qu.modifiers[ErrorTimeField][0]
	if list.vars[three.(*Filter).Variables[1].Value] != "filterOne" {
		t.Fail()
		return
	}
	four := q.modifiers[ErrorMessageField][0]
	if list.vars[four.(*Filter).Variables[1].Value] != "5" {
		t.Fail()
		return
	}
}

//should average about 4 µs per query generation and 40 allocations.
//can it be better?
func BenchmarkQueries(b *testing.B) {
	for i := 0; i < b.N; i++ {
		list := NewQueries()
		q := list.NewQuery(ErrorFields)
		q.Function(Equals).PredValue(ErrorTimeField, "testFunctionOne")
		q.Filter(MakeFilter(Equals).PredValue(ErrorTimeField, "testFilterOne"), "")
		q.Order(Ascending, "", ErrorTimeField)
		q.Count(CountFirst, "", 5)
		qu := list.NewQuery(ErrorFields)
		qu.Function(Equals).PredValue(ErrorTimeField, "testFunctionTwo")
		qu.Filter(MakeFilter(Equals).PredValue(ErrorTimeField, "testFilterOne"), "")
		qu.Order(Ascending, "", ErrorTimeField)
		qu.Count(CountFirst, "", 5)
		_, _ = list.Process()
	}
	b.ReportAllocs()
}

func TestVariable(t *testing.T) {
	const expected = "query{q(func: has(<Error.message>),first:5,orderasc: Error.time){Error.message Error.errorType Error.time  test : math(p)  p as ErrorTimeField  uid}}"
	q := NewQuery(ErrorFields)
	q.Variable("test", "", "math(p)", true)
	q.Variable("p", "", "ErrorTimeField", false)
	q.Function(Has).Pred(ErrorMessageField)
	//q.Filter(MakeFilter(Equals).PredValue(ErrorTimeField, "testFilter"), "")
	q.Order(Ascending, "", ErrorTimeField)
	q.Count(CountFirst, "", 5)
	str, err := q.Process()
	if err != nil {
		t.Fail()
		return
	}
	if str != expected {
		t.Fail()
		return
	}
}

func TestMultipleUid(t *testing.T) {
	q := NewQuery(ErrorFields)
	q.Function(FunctionUid).Values("0x1", "0x2")

	str, _ := q.Process()
	fmt.Println(str)
}

var a DNode = &dbError{}

func BenchmarkDNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		x := a.(DNode)
		_ = x
	}
}

func TestWithFunction(t *testing.T) {
	q := NewQuery(ErrorFields)
	q.Function(Less.WithFunction("val")).Values("name", 1)

	s, _ := q.Process()
	fmt.Println(s)
}
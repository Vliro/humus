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

var fields = ErrorFields.Sub(ErrorMessageField, ErrorFields)

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

func TestAt(t *testing.T) {
	q := NewQuery(ErrorFields).Function(Equals).Values(ErrorMessageField, "yolo")

	q.At("", func(m Mod) {
		m.Sort(Descending, ErrorMessageField)
		m.Sort(Ascending, ErrorErrorTypeField)
		m.Filter(FunctionUid, "0x1", "0x2")
	})

	q.GroupBy(ErrorTimeField, ErrorMessageField, func(m Mod) {
		m.Variable("yolo", "swag", false)
		m.Variable("hej", "swag", false)
		m.Variable("cool", "swag", false)
	})

	q.Facets(ErrorErrorTypeField, func(m Mod) {
		m.Variable("test", "test", false)
		m.Variable("d", "d", false)
	})

	str, err := q.Process()
	print(err)
	fmt.Println(str)
}

func BenchmarkAt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := NewQuery(ErrorFields).Function(Equals).Values(ErrorMessageField, "yolo")

		q.At("", func(m Mod) {
			m.Sort(Descending, ErrorMessageField)
			m.Sort(Ascending, ErrorErrorTypeField)
			m.Filter(FunctionUid, "0x1", "0x2")
		})
		/*
			q.GroupBy(ErrorTimeField, ErrorMessageField, func(m Mod) {
				m.Variable("yolo", "swag", false)
				m.Variable("hej", "swag", false)
				m.Variable("cool", "swag", false)
			})

			q.Facets(ErrorErrorTypeField, func(m Mod) {
				m.Variable("test", "test", false)
				m.Variable("d", "d", false)
			})*/

		_, _ = q.Process()
	}
	b.ReportAllocs()
}

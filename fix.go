package mulbase

import "fmt"

type StaticQuery struct {
	Query string
}

func (s StaticQuery) Type() QueryType {
	return QueryRegular
}

func (s StaticQuery) Process(sch schemaList) ([]byte, map[string]string, error) {
	return []byte(s.Query), nil, nil
}

func NewStaticQuery(query string, vals ...interface{}) StaticQuery {
	str := fmt.Sprintf(query, vals...)
	return StaticQuery{str}
}

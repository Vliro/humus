package mulbase

import "fmt"

//StaticQuery represents a static query.
type StaticQuery struct {
	Query []byte
}

func (s StaticQuery) Type() QueryType {
	return QueryRegular
}

func (s StaticQuery) Process(sch schemaList) ([]byte, map[string]string, error) {
	//TODO: API Forces copying here. Should we change it up?
	//TODO: Add GraphQL variables.
	return s.Query, nil, nil
}

func NewStaticQuery(query string, vals ...interface{}) StaticQuery {
	str := fmt.Sprintf(query, vals...)
	return StaticQuery{[]byte(str)}
}

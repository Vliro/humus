package mulbase

//StaticQuery represents a static query.
type StaticQuery struct {
	Query []byte
	vars map[string]string
}
//SetVar sets the variable in the GraphQL query map.
func (s *StaticQuery) SetVar(key string, val interface{}) *StaticQuery {
	if s.vars == nil {
		s.vars = make(map[string]string)
	}
	s.vars[key],_ = processInterface(val)
	return s
}
//Type satisfies the interface Query.
func (s StaticQuery) Type() QueryType {
	return QueryRegular
}
//Process the query in order to send to DGraph.
func (s StaticQuery) Process(sch SchemaList) ([]byte, map[string]string, error) {
	//TODO: API Forces copying here. Should we change it up?
	//TODO: Add GraphQL variables.
	return s.Query, s.vars, nil
}
//NewStaticQuery creates a formatted query. Use
func NewStaticQuery(query string) StaticQuery {
	return StaticQuery{Query:[]byte(query)}
}

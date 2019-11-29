package mulbase

//StaticQuery represents a static query.
type StaticQuery struct {
	Query string
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

//Process the query in order to send to DGraph.
func (s StaticQuery) Process(sch SchemaList) (string, error) {
	//TODO: API Forces copying here. Should we change it up?
	//TODO: Add GraphQL variables.
	return s.Query, nil
}

func (s StaticQuery) Vars() map[string]string {
	return s.vars
}

//NewStaticQuery creates a formatted query. Use
func NewStaticQuery(query string) StaticQuery {
	return StaticQuery{Query:query}
}
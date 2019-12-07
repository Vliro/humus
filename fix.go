package humus

//StaticQuery represents a static query.
type StaticQuery struct {
	Query string
	vars map[string]string
}
//SetVar sets the variable in the GraphQL query map.
//Since it is a static query the type is expected to be supplied
//in the string part of the query and not here.
func (s *StaticQuery) SetVar(key string, val interface{}) *StaticQuery {
	if s.vars == nil {
		s.vars = make(map[string]string)
	}
	s.vars[key],_ = processInterface(val)
	return s
}

//Process the query in order to send to DGraph.
func (s StaticQuery) Process() (string, error) {
	return s.Query, nil
}

func (s StaticQuery) Vars() map[string]string {
	return s.vars
}

//NewStaticQuery creates a formatted query.
//Use fmt.Sprintf as well as SetVar to supply parameters.
func NewStaticQuery(query string) StaticQuery {
	return StaticQuery{Query:query}
}
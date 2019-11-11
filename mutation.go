package mulbase

type SingleMutation struct {
	Object interface{}
	QueryType QueryType
}

func (m SingleMutation) Type() QueryType {
	if m.QueryType == QueryRegular {
		panic("mutation: set to invalid query type.")
	}
	return m.QueryType
}

func checkInterface() {

}

func (m SingleMutation) Process(schemaList) ([]byte, map[string]string, error) {
	//panic("do not process a single mutation")
	b, _ := json.Marshal(m.Object)
	return b, nil, nil
}

type MutationQuery struct {
	Values []SingleMutation
	Type MutationType
}

func (m *MutationQuery) Process(list schemaList) (string, map[string]string, error) {
	return "", nil, nil
}



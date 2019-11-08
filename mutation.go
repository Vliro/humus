package mulbase

type SingleMutation struct {
	Object interface{}
	QueryType QueryType
}

func (m SingleMutation) Type() QueryType {
	return m.QueryType
}

func (m SingleMutation) Process() ([]byte, map[string]string, error) {
	//panic("do not process a single mutation")
	b, _ := json.Marshal(m.Object)
	return b, nil, nil
}

type MutationQuery struct {
	Values []SingleMutation
	Type MutationType
}

func (m *MutationQuery) Process() (string, map[string]string, error) {
	return "", nil, nil
}



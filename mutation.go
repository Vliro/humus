package mulbase

//SingleMutation represents just that, one object mutated.
//interface{} is used over DNode as the structure of a mutation might
//change so a map[string]interface{} is needed for certain mutations.
type SingleMutation struct {
	Object interface{}
	QueryType QueryType
}

func (m SingleMutation) Type() QueryType {
	return m.QueryType
}

func (m SingleMutation) Process(schemaList) ([]byte, map[string]string, error) {
	//panic("do not process a single mutation")
	b, _ := json.Marshal(m.Object)
	return b, nil, nil
}

type MutationQuery struct {
	Values []interface{}
	QueryType QueryType
}

func (m *MutationQuery) Process(list schemaList) ([]byte, map[string]string, error) {
	byt, err := json.Marshal(m.Values)
	return byt, nil, err
}

func (m *MutationQuery) Type() QueryType {
	return m.QueryType
}

package mulbase

type SingleMutation struct {
	Object interface{}
	Type MutationType
}

func (m SingleMutation) Process() (string, map[string]string, error) {
	panic("do not process a single mutation")
	//b, _ := json.Marshal(m.Object)
	//return bytesToStringUnsafe(b), nil, nil
}

type MutationQuery struct {
	Values []SingleMutation
	Type MutationType
}

func (m *MutationQuery) Process() (string, map[string]string, error) {
	return "", nil, nil
}



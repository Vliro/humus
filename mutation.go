package mulbase

import (
	"encoding/json"
	"github.com/pkg/errors"
)

//SingleMutation represents just that, one object mutated.
//interface{} is used over DNode as the structure of a mutation might
//change so a map[string]interface{} is needed for certain mutations.
type SingleMutation struct {
	Object DNode
	QueryType QueryType
}

func (m SingleMutation) Type() QueryType {
	return m.QueryType
}

func (m SingleMutation) Process(schemaList) ([]byte, map[string]string, error) {
	//panic("do not process a single mutation")
	if m.Object == nil {
		return nil, nil, errors.New("nil value supplied to process")
	}
	var b []byte
	if val, ok := m.Object.(Saver); ok {
		if val == nil {
			return nil, nil, errors.New("nil value supplied to process")
		}
		b, _ = json.Marshal(val.Save())
	} else {
		b, _ = json.Marshal(m.Object)
	}
	return b, nil, nil
}

type MutationQuery struct {
	Values []interface{}
	QueryType QueryType
}

func (m *MutationQuery) Process(list schemaList) ([]byte, map[string]string, error) {
	for k,v := range m.Values {
		if val, ok := v.(Saver); ok {
			if val == nil {
				return nil, nil, errors.New("nil value supplied to process")
			}
			m.Values[k] = val.Save()
		}
	}
	byt, err := json.Marshal(m.Values)
	return byt, nil, err
}

func (m *MutationQuery) Type() QueryType {
	return m.QueryType
}

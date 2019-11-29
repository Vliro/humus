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
	MutationType MutationType
}

func (m SingleMutation) Type() MutationType {
	return m.MutationType
}

func (m SingleMutation) Mutate() ([]byte, error) {
	//panic("do not process a single mutation")
	if m.Object == nil {
		return nil, errors.New("nil value supplied to process")
	}
	if m.Object.UID() == "" {
		m.Object.SetType()
	}
	var b []byte
	switch m.MutationType {
	case MutateSet:
		if val, ok := m.Object.(Saver); ok {
			if val == nil {
				return nil, errors.New("nil value supplied to process")
			}
			b, _ = json.Marshal(val.Save())
		} else {
			b, _ = json.Marshal(m.Object)
		}
	case MutateDelete:
		if val, ok := m.Object.(Deleter); ok {
			if val == nil {
				return nil, errors.New("nil value supplied to process")
			}
			b, _ = json.Marshal(val.Delete())
		} else {
			b, _ = json.Marshal(m.Object)
		}
	}
	return b, nil
}

type MutationQuery struct {
	Values []DNode
	MutationType MutationType
}

func (m *MutationQuery) Mutate() ([]byte, error) {
	for k,v := range m.Values {
		if v.UID() == "" {
			v.SetType()
		}
		switch m.MutationType {
		case MutateSet:
			if val, ok := v.(Saver); ok {
				if val == nil {
					return nil, errors.New("nil value supplied to process")
				}
				m.Values[k] = val.Save()
			}
		case MutateDelete:
			if val, ok := v.(Deleter); ok {
				if val == nil {
					return nil, errors.New("nil value supplied to process")
				}
				m.Values[k] = val.Delete()
			}
		}
	}
	byt, err := json.Marshal(m.Values)
	return byt, err
}

func (m *MutationQuery) Type() MutationType {
	return m.MutationType
}

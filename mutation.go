package humus

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
	//Used for upsert.
	Condition string
}

func (m SingleMutation) Type() MutationType {
	return m.MutationType
}

func (m SingleMutation) Cond() string {
	return m.Condition
}

func (m SingleMutation) mutate() ([]byte, error) {
	//panic("do not process a single mutation")
	if m.Object == nil {
		return nil, errors.New("nil value supplied to process")
	}
	m.Object.Recurse()
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
	Condition string
	MutationType MutationType
}
//CreateMutations creates a list of mutations from a variadic list of Dnodes.
func CreateMutations(typ MutationType, muts ...DNode) *MutationQuery {
	return &MutationQuery{
		Values:       muts,
		MutationType: typ,
	}
}

func (m *MutationQuery) SetCondition(c string) *MutationQuery {
	m.Condition = c
	return m
}

func (m *MutationQuery) Cond() string {
	return m.Condition
}

func (m *MutationQuery) mutate() ([]byte, error) {
	for k,v := range m.Values {
		v.Recurse()
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

type customMutation struct {
	Value interface{}
	QueryType MutationType
	Condition string
}

func (c customMutation) Cond() string {
	return c.Condition
}

func (c customMutation) mutate() ([]byte, error) {
	b, _ := json.Marshal(c.Value)
	return b, nil
}

func (c customMutation) Type() MutationType {
	return c.QueryType
}
//CreateCustomMutation allows you to create a mutation from an interface
//and not a DNode. This is useful alongside custom queries to set values, especially
//working with facets.
func CreateCustomMutation(obj interface{}, typ MutationType) Mutate {
	return customMutation{
		Value: obj,
		QueryType:  typ,
	}
}
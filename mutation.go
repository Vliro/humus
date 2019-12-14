package humus

import (
	"github.com/pkg/errors"
)

//SingleMutation represents just that, one object mutated.
//interface{} is used over DNode as the structure of a mutation might
//change so a map[string]interface{} is needed for certain mutations.
type SingleMutation struct {
	Object       DNode
	MutationType MutationType
	//Used for upsert.
	Condition string
}

func (s SingleMutation) WithCond(cond string) SingleMutation {
	s.Condition = cond
	return s
}

func (s SingleMutation) Type() MutationType {
	return s.MutationType
}

func (s SingleMutation) Cond() string {
	return s.Condition
}

func (s SingleMutation) mutate() ([]byte, error) {
	//panic("do not Process a single mutation")
	if s.Object == nil {
		return nil, errors.New("nil value supplied to Process")
	}
	s.Object.Recurse(0)
	var b []byte
	switch s.MutationType {
	case MutateSet:
		if val, ok := s.Object.(Saver); ok {
			if val == nil {
				return nil, errors.New("nil value supplied to Process")
			}
			node, err := val.Save()
			if err != nil {
				return nil, err
			}
			b, _ = json.Marshal(node)
		} else {
			b, _ = json.Marshal(s.Object)
		}
	case MutateDelete:
		if val, ok := s.Object.(Deleter); ok {
			if val == nil {
				return nil, errors.New("nil value supplied to Process")
			}
			node, err := val.Delete()
			if err != nil {
				return nil, err
			}
			b, _ = json.Marshal(node)
		} else {
			b, _ = json.Marshal(s.Object)
		}
	}
	return b, nil
}

type MutationQuery struct {
	Values       []DNode
	Condition    string
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
	var counter int
	var err error
	for k, v := range m.Values {
		counter = v.Recurse(counter)
		switch m.MutationType {
		case MutateSet:
			if val, ok := v.(Saver); ok {
				if val == nil {
					return nil, errors.New("nil DNode supplied to Process in mutationQuery")
				}
				m.Values[k], err = val.Save()
				if err != nil {
					return nil, err
				}
			}
		case MutateDelete:
			if val, ok := v.(Deleter); ok {
				if val == nil {
					return nil, errors.New("nil DNode supplied to Process in mutationQuery")
				}
				m.Values[k], err = val.Delete()
				if err != nil {
					return nil, err
				}
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
	Value     interface{}
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
		Value:     obj,
		QueryType: typ,
	}
}

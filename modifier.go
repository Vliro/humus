package humus

import (
	"errors"
	"strings"
)

type modifierSource int

/*
Begin simple mods
*/

type mapElement struct {
	m modifierList
	g groupBy
	f facet
	q *GeneratedQuery
}

type facetCreator mapElement

type groupCreator mapElement

type modifierCreator mapElement

const (
	modifierField modifierSource = iota
	modifierFunction
)

/*
Operation is a closure callback given a mod. Any operations called on this
applies the given operation at the path.
*/
type Operation func(m Mod)

/*
Mod is the core for applying operations at certain predicate levels.
There are two kind of 'mods', those which exist at root/edge level and those
at field level. Paginate, Filter, Sort exists at root/edge level with the remaining
existing at field level. This is an important distinction
For Paginate, a path of "" applies the pagination at the top level (root) and given a single
predicate P it applies it on the edge.
For Variable, a path of "" applies the variable at the root field level, at the top level node.
*/
type Mod interface {
	/*
		Paginate creates a pagination at this level given the pagination type
		and the value.
	*/
	Paginate(t PaginationType, value int) bool
	/*
		Filter creates a filter at this level given a function type and a list of variables with the
		same syntax as a function.
	*/
	Filter(t FunctionType, variables ...interface{}) bool
	/*
		Sort applies a sorting at this level.
	*/
	Sort(t OrderType, p Predicate) bool
	/*
		Aggregate sets an aggregation at this level.
	*/
	Aggregate(t AggregateType, v string, alias string) bool
	/*
		Count sets a count variable at this level, e.g.
		result : count(uid) given a "uid" as predicate and
		"result" as alias.
	*/
	Count(p Predicate, alias string) bool
	/*
		Variable sets a variable at this level.
		It either generates a value variable or an alias variable.
		If name is omitted so is the prefix for the variable. This
		can be useful for setting facet variables where name is omitted.
	*/
	Variable(name string, value string, isAlias bool) bool
}

type modifierType uint8

const (
	modifierVariable modifierType = 1 << iota
	modifierAggregate
	//All above are field generating modifiers.
	modifierFilter
	modifierPagination
	modifierOrder
	modifierGroupBy
	modifierFacet
)

type modifier interface {
	canApply(mt modifierSource) bool
	//While io.Writer is more generic, the utility of
	//multiple different write methods is unbeatable here.
	apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error
	priority() modifierType
	parenthesis() bool
}

type modifierList []modifier

func (m modifierList) runNormal(q *GeneratedQuery, meta FieldMeta, where modifierSource, sb *strings.Builder) error {
	var curType modifierType
	for k, v := range m {
		newType := v.priority()
		if newType <= modifierAggregate {
			continue
		}
		if v.canApply(where) {
			if newType > curType && v.parenthesis() {
				sb.WriteByte('(')
			}
			err := v.apply(q, meta, where, sb)
			if k != len(m)-1 && v.parenthesis() {
				if p := m[k+1].priority(); p == newType {
					sb.WriteByte(',')
				} else if p != newType {
					sb.WriteByte(')')
				}
			}
			if err != nil {
				return err
			}
		}
		if k == len(m)-1 && v.parenthesis() {
			sb.WriteByte(')')
		}
		curType = newType
	}
	return nil
}

func (m modifierList) runTopLevel(q *GeneratedQuery, meta FieldMeta, where modifierSource, sb *strings.Builder) error {
	for _, v := range m {
		if v.priority() > modifierFilter && v.canApply(modifierFunction) {
			sb.WriteByte(',')
			err := v.apply(q, 0, modifierFunction, sb)
			if err != nil {
				return err
			}
		}
	}
	sb.WriteByte(')')
	for _, v := range m {
		if v.priority() == modifierFilter && v.canApply(modifierFunction) {
			err := v.apply(q, 0, modifierFunction, sb)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m modifierList) runVariables(q *GeneratedQuery, meta FieldMeta, where modifierSource, sb *strings.Builder, commaSeparated bool) error {
	for k, v := range m {
		if v.priority() > modifierAggregate {
			break
		}
		err := v.apply(q, meta, where, sb)
		if err != nil {
			return err
		}
		if commaSeparated && k < len(m)-1 {
			sb.WriteByte(',')
		}
	}
	return nil
}

func (m *modifierCreator) Paginate(t PaginationType, value int) bool {
	m.m = append(m.m, pagination{Type: t, Value: value})
	return true
}

func (m *modifierCreator) Variable(name string, value string, isAlias bool) bool {
	m.m = append(m.m, variable{
		name:  name,
		value: value,
		alias: isAlias,
	})
	return true
}

func (m *modifierCreator) Filter(t FunctionType, variables ...interface{}) bool {
	var filter Filter
	filter.typ = t
	filter.variables = make([]graphVariable, len(variables))
	for k, v := range variables {
		val, typ := processInterface(v)
		filter.variables[k] = graphVariable{
			Value: val,
			Type:  typ,
		}
	}
	filter.mapVariables(m.q)
	m.m = append(m.m, &filter)
	return true
}

func (m *modifierCreator) Sort(t OrderType, p Predicate) bool {
	m.m = append(m.m, Ordering{
		Type:      t,
		Predicate: p,
	})
	return true
}

func (m *modifierCreator) Aggregate(t AggregateType, v string, alias string) bool {
	m.m = append(m.m, aggregateValues{
		Type:     t,
		Alias:    alias,
		Variable: v,
	})
	return true
}

func (m *modifierCreator) Count(p Predicate, alias string) bool {
	m.m = append(m.m, variable{
		name:  alias,
		value: "count(" + string(p) + ")",
		alias: true,
	})
	return true
}

func (m modifierList) Len() int {
	return len(m)
}

func (m modifierList) hasModifier(mt modifierType) bool {
	for _, v := range m {
		if v.priority() == mt {
			return true
		}
	}
	return false
}

func (m modifierList) Less(i, j int) bool {
	return m[i].priority() < m[j].priority()
}

func (m modifierList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

//aggregateValues represents a modifier with a type(sum),
//an alias for changing json key as well as what variable or predicate it acts on.
type aggregateValues struct {
	Type     AggregateType
	Alias    string
	Variable string
}

func (a aggregateValues) canApply(mt modifierSource) bool {
	return true
}

func (a aggregateValues) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	sb.WriteByte(' ')
	if a.Variable == "" {
		return errors.New("missing predicate in aggregateValues")
	}
	if a.Alias != "" {
		sb.WriteString(a.Alias)
		sb.WriteString(" : ")
	}
	sb.WriteString(string(a.Type))
	isCount := a.Type == "count"
	sb.WriteByte('(')
	if !isCount {
		sb.WriteString("val(")
	}
	sb.WriteString(a.Variable)
	if !isCount {
		sb.WriteByte(')')
	}
	sb.WriteByte(')')
	sb.WriteByte(' ')
	return nil
}

func (a aggregateValues) priority() modifierType {
	return modifierAggregate
}

func (a aggregateValues) parenthesis() bool {
	return true
}

type groupBy struct {
	m modifierList
	p Predicate
}

func (g *groupCreator) Paginate(t PaginationType, value int) bool {
	return false
}

func (g *groupCreator) Variable(name string, value string, isAlias bool) bool {
	g.g.m = append(g.g.m, variable{
		name:  name,
		value: value,
		alias: isAlias,
	})
	return true
}

func (g *groupCreator) Filter(t FunctionType, variables ...interface{}) bool {
	return false
}

func (g *groupCreator) Sort(t OrderType, p Predicate) bool {
	return false
}

func (g *groupCreator) Aggregate(t AggregateType, v string, alias string) bool {
	g.g.m = append(g.g.m, aggregateValues{
		Type:     t,
		Alias:    alias,
		Variable: v,
	})
	return true
}

func (g *groupCreator) Count(p Predicate, alias string) bool {
	return false
}

func (g groupBy) canApply(mt modifierSource) bool {
	return mt == modifierField
}

func (g groupBy) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	if g.p == "" {
		return errors.New("missing predicate type in groupBy")
	}
	sb.WriteString("@groupby(")
	sb.WriteString(string(g.p))
	sb.WriteByte(')')
	sb.WriteByte('{')
	g.m.runVariables(root, 0, mt, sb, false)
	sb.WriteByte('}')
	return nil
}

func (g groupBy) priority() modifierType {
	return modifierGroupBy
}

func (g groupBy) parenthesis() bool {
	return false
}

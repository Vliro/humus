package humus

import (
	"fmt"
	"strings"
)

type FunctionType string

//WithFunction creates an in/equality function with a subfunction.
//Possible values for typ is count and val. For instance, cases with
//lt(count(predicate),1). would be Less.WithFunction("count") with two function variables,
//predicate, 1.
func (f FunctionType) WithFunction(typ string) FunctionType {
	return FunctionType(string(f) + "(" + typ)
}

type OrderType string

type Predicate string

//Stringify this predicate.
func (s Predicate) String() string {
	if s[len(s)-1] == ' ' {
		return string(s[:len(s)-1])
	}
	return string(s)
}

const (
	Equals      FunctionType = "eq"
	AllOfText   FunctionType = "alloftext"
	AllOfTerms  FunctionType = "allofterms"
	AnyOfTerms  FunctionType = "anyofterms"
	FunctionUid FunctionType = "uid"
	Has         FunctionType = "has"
	LessEq      FunctionType = "le"
	Match       FunctionType = "match"
	Less        FunctionType = "lt"
	GreaterEq   FunctionType = "ge"
	Greater     FunctionType = "gt"
	Type        FunctionType = "type"
)

const (
	Ascending  OrderType = "orderasc"
	Descending OrderType = "orderdesc"
)

type Ordering struct {
	Type      OrderType
	Predicate Predicate
}

func (o Ordering) parenthesis() bool {
	return true
}

func (o Ordering) canApply(mt modifierSource) bool {
	return true
}

func (o Ordering) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, s *strings.Builder) error {
	s.WriteString(string(o.Type))
	s.WriteString(": ")
	s.WriteString(o.Predicate.String())
	return nil
}

func (o Ordering) priority() modifierType {
	return modifierOrder
}

//If you change the names here make sure to follow the type typ + name as dictated in gen.
type varType string

const (
	typeString  varType = "string"
	typeInt     varType = "int"
	typePred    varType = "ignore"
	typeUid     varType = "uid"
	typeFloat   varType = "float"
	typeGeo     varType = "geo"
	typeDefault varType = ""
)

//graphVariable represents a variable before it is parsed and written into a query.
type graphVariable struct {
	Value string
	Type  varType
}

//function represents a GraphQL+- function. It writes into the query
//the function type, checks the type as well as the list of arguments.
//It uses GraphQL variables to minimize risk of any type of injection.
type function struct {
	typ       FunctionType
	variables []graphVariable
}

//OR, NOT, AND implements basic logicals.
//This is not a fully featured function, made for simple examples such as single and or single OR
//Not working right now. For complex filters just use fix queries for now.

/*
func (f *function) OR(fi *Filter) *Filter {
	return &Filter{
		op:            logicalOr,
		function: *f,
		node:          fi,
	}
}

func (f *function) NOT(fi *Filter) *Filter {
	return &Filter{
		op:            logicalNot,
		function: *f,
		node:          fi,
	}
}

func (f *function) AND(fi *Filter) *Filter {
	return &Filter{
		op:            logicalAnd,
		function: *f,
		node:          fi,
	}
}
*/
//Functions returns a new function. It preallocates a list of size four, a common case most likely.
func newFunction(ft FunctionType) function {
	return function{typ: ft, variables: make([]graphVariable, 0, 4)}
}

//value adds a variable to the function depending on its type.
//Possible values are int, float, string, predicate, uid.
//Other values are possible but will be formatted as fmt.Sprintf.
func (f *function) value(v interface{}) *function {
	val, typ := processInterface(v)
	vv := graphVariable{val, typ}
	f.variables = append(f.variables, vv)
	return f
}

func (f *function) values(v []interface{}) *function {
	for k := range v {
		val, typ := processInterface(v[k])
		f.variables = append(f.variables, graphVariable{val, typ})
	}
	return f
}

func (f *function) pred(name Predicate) *function {
	f.variables = append(f.variables, graphVariable{string(name), typePred})
	return f
}

//PredValue is simply a shorthand for function such as equals.
func (f *function) predValue(name Predicate, v interface{}) *function {
	val, typ := processInterface(v)
	v1 := graphVariable{string(name), typePred}
	v2 := graphVariable{val, typ}
	f.variables = append(f.variables, v1, v2)
	return f
}

func (f *function) predMultiple(name Predicate, v []interface{}) *function {
	v1 := graphVariable{string(name), typePred}
	f.variables = append(f.variables, v1)
	for _, vv := range v {
		val, typ := processInterface(vv)
		v2 := graphVariable{val, typ}
		f.variables = append(f.variables, v2)
	}
	return f
}

func (f *function) mapVariables(q *GeneratedQuery) {
	for k, v := range f.variables {
		//Handle the two special cases that do not need variable mapping.
		if v.Type == typePred {
			continue
		}
		if v.Type == typeUid {
			if len(v.Value) > 16 {
				panic("invalid UID, this could be an SQL injection.")
			}
			continue
		}
		//Do not cause variable names in a function to be GraphQL mapped.
		if k == 0 && strings.IndexByte(string(f.typ), '(') != -1 {
			continue
		}
		//Build the variable using the integer from the query.
		key := q.registerVariable(v.Type, v.Value)
		f.variables[k].Value = key
	}
}

func (f *function) create(q *GeneratedQuery, sb *strings.Builder) error {
	if err := f.check(q); err != nil {
		return err
	}
	//f.mapVariables(q)
	//Write the default values.
	sb.WriteString(string(f.typ))
	sb.WriteByte('(')
	f.buildVarString(sb)
	sb.WriteByte(')')
	return nil
}

func (f *function) buildVarString(sb *strings.Builder) {
	for k, v := range f.variables {
		switch v.Type {
		case typePred:
			sb.WriteByte('<')
			sb.WriteString(Predicate(v.Value).String())
			sb.WriteByte('>')
			/*
				Handle custom functions.
			*/
		case typeUid:
			sb.WriteByte('"')
			sb.WriteString(v.Value)
			sb.WriteByte('"')
		default:
			sb.WriteString(v.Value)
		}
		if k == 0 && strings.IndexByte(string(f.typ), '(') != -1 {
			sb.WriteByte(')')
		}
		if k != len(f.variables)-1 {
			sb.WriteByte(',')
		}
	}
}

func (f *function) check(q *GeneratedQuery) error {
	if f.typ == "" {
		return errMissingFunction
	}
	if len(f.variables) == 0 {
		return errMissingVariables
	}
	switch f.typ {
	case Has:
		if len(f.variables) != 1 && f.variables[0].Type != typePred {
			return fmt.Errorf("%s function too many variables or invalid types, have %v need %v", f.typ, len(f.variables), 1)
		}
		break
	}
	return nil
}

package mulbase

import (
	"fmt"
	"strconv"
	"strings"
)

type functionType string
type OrderType string

type Predicate string

func (s Predicate) Sub(name Predicate, fields Fields) Fields {
	panic("implement me")
}

func (s Predicate) Add(fi Field) Fields {
	panic("implement me")
}

func (s Predicate) Facet(facetName string, alias string) Fields {
	panic("implement me")
}

func (s Predicate) Get() []Field {
	return nil
}

func (s Predicate) Len() int {
	return 1
}

//Stringify this predicate.
func (s Predicate) String() string {
	if s[len(s)-1] == ' ' {
		return string(s[:len(s)-1])
	}
	return string(s)
}

const (
	FunctionEquals     functionType = "eq"
	FunctionAllOfText  functionType = "alloftext"
	FunctionAllOfTerms functionType = "allofterms"
	FunctionAnyOfTerms functionType = "anyofterms"
	FunctionUid        functionType = "uid"
	FunctionHas        functionType = "has"
	FunctionLessEq     functionType = "le"
	FunctionMatch      functionType = "match"
	FunctionLess       functionType = "lt"
	FunctionGreaterEq  functionType = "ge"
	FunctionGreater    functionType = "gt"
	FunctionType       functionType = "type"
)

const (
	OrderAsc  OrderType = "orderasc"
	OrderDesc OrderType = "orderdesc"
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

func (o Ordering) apply(root *GeneratedQuery, meta FieldMeta, name string, s *strings.Builder) (modifierType, error) {
	s.WriteString(string(o.Type))
	s.WriteString(": ")
	s.WriteString(o.Predicate.String())
	return 0, nil
}

func (o Ordering) priority() modifierType {
	return modifierOrder
}

/*func (o Ordering) String() string {
	var s strings.Builder
	s.WriteString(string(o.Type))
	s.WriteString(": ")
	s.WriteString(o.Predicate.String())
	s.WriteByte(' ')
	return s.String()
}*/

//If you change the names here make sure to follow the type Type + name as dictated in gen.
type VarType string

const (
	TypeStr     VarType = "string"
	TypeInt     VarType = "int"
	TypePred    VarType = "ignore"
	TypeUid     VarType = "uid"
	TypeDefault VarType = ""
)

//GraphVariable represents a variable before it is parsed and written into a query.
type GraphVariable struct {
	Value string
	Type  VarType
}

//Function represents a GraphQL+- function. It writes into the query
//the function type, checks the type as well as the list of arguments.
//It uses GraphQL variables to minimize risk of any type of injection.
type Function struct {
	Type functionType
	//Is it lazy to use []interface? yes!
	//but then you dont have to specify variable types :)
	Variables []GraphVariable
}

func (f *Function) canApply(mt modifierSource) bool {
	return true
}

func (f *Function) Apply(root *GeneratedQuery, meta FieldMeta, w *strings.Builder) error {
	err := f.create(root, w)
	return err
}

func MakeFunction(ft functionType) *Function {
	return &Function{Type: ft, Variables: make([]GraphVariable, 0, 2)}
}

func (f *Function) AddOrdering(t OrderType, pred string) *Function {
	//f.Order = append(f.Order, Ordering{Type: t, Predicate: Predicate(pred)})
	return f
}

func (f *Function) AddValue(v interface{}) *Function {
	val, typ := processInterface(v)
	vv := GraphVariable{val, typ}
	f.Variables = append(f.Variables, vv)
	return f
}

func (f *Function) AddValues(v ...interface{}) *Function {
	for k := range v {
		val, typ := processInterface(v[k])
		v2 := GraphVariable{val, typ}
		f.Variables = append(f.Variables, v2)
	}
	return f
}

func (f *Function) AddPred(name Predicate) *Function {
	vv := GraphVariable{string(name), TypePred}
	f.Variables = append(f.Variables, vv)
	return f
}

func (f *Function) AddPredValue(name Predicate, v interface{}) *Function {
	val, typ := processInterface(v)
	v1 := GraphVariable{string(name), TypePred}
	v2 := GraphVariable{val, typ}
	f.Variables = append(f.Variables, v1, v2)
	return f
}

func (f *Function) AddMatchValues(name Predicate, v string, count int) *Function {
	v1 := GraphVariable{string(name), TypePred}
	v2 := GraphVariable{v, TypeStr}
	v3 := GraphVariable{strconv.Itoa(count), TypeInt}
	f.Variables = append(f.Variables, v1, v2, v3)
	return f
}

func (f *Function) AddPredMultiple(name Predicate, v ...interface{}) *Function {
	v1 := GraphVariable{string(name), TypePred}
	f.Variables = append(f.Variables, v1)
	for _, vv := range v {
		val, typ := processInterface(vv)
		v2 := GraphVariable{val, typ}
		f.Variables = append(f.Variables, v2)
	}
	return f
}

func (f *Function) mapVariables(q *GeneratedQuery) {
	for k, v := range f.Variables {
		//Handle the two special cases that do not need variable mapping.
		if v.Type == TypePred {
			continue
		}
		if v.Type == TypeUid {
			if len(v.Value) > 16 {
				panic("invalid UID, this could be an SQL injection.")
			}
			continue
		}
		//Build the variable using the integer from the query.
		key := q.registerVariable(v.Type, v.Value)
		f.Variables[k].Value = key
	}
}

func (f *Function) create(q *GeneratedQuery, sb *strings.Builder) error {
	if err := f.check(q); err != nil {
		return err
	}
	//Map the variables to their proper value.
	f.mapVariables(q)
	//Write the default values.
	sb.WriteString(string(f.Type))
	sb.WriteString(tokenLP)
	f.buildVarString(sb)
	sb.WriteString(tokenRP)
	/*for k := range f.Order {
		sb.WriteString(tokenComma)
		s := f.Order[k].String()
		sb.WriteString(s)
	}*/
	return nil
}

func (f *Function) buildVarString(sb *strings.Builder) {
	for k, v := range f.Variables {
		if v.Type == TypePred {
			sb.WriteByte('<')
			sb.WriteString(Predicate(v.Value).String())
			sb.WriteByte('>')
		}else if v.Type == TypeUid {
			sb.WriteByte('"')
			sb.WriteString(v.Value)
			sb.WriteByte('"')
		} else {
			sb.WriteString(v.Value)
		}
		if k != len(f.Variables)-1 {
			sb.WriteByte(',')
		}
	}
}

type FunctionError struct {
	Type functionType
}

func (f FunctionError) Error() string {
	return fmt.Sprintf("invalid arguments of function type %s", string(f.Type))
}
func NewFunctionError(t functionType) FunctionError {
	return FunctionError{t}
}

func (f *Function) check(q *GeneratedQuery) error {
	if f == nil {
		return errMissingFunction
	}
	if f.Type == "" {
		return errMissingFunction
	}
	if len(f.Variables) == 0 {
		return errMissingVariables
	}
	switch f.Type {
	case FunctionHas:
		if len(f.Variables) != 1 && f.Variables[0].Type != TypePred {
			return fmt.Errorf("%s function too many variables or invalid types, have %v need %v", f.Type, len(f.Variables), 1)
		}
		break
	}
	return nil
}

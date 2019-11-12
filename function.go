package mulbase

import (
	"bytes"
	"fmt"
	"strings"
)

type functionType string
type OrderType string

type Predicate string
//Stringify this predicate.
func (s Predicate) String() string {
	return string("<" + s + ">")
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

func (o Ordering) String() string {
	var s strings.Builder
	s.WriteString(string(o.Type))
	s.WriteString(": ")
	s.WriteString(o.Predicate.String())
	s.WriteByte(' ')
	return s.String()
}

//If you change the names here make sure to follow the type Type + name as dictated in mulgen.
type VarType string

const (
	TypeStr     VarType = "string"
	TypeInt     VarType = "int"
	TypePred    VarType = "ignore"
	TypeUid     VarType = "uid"
	TypeDefault VarType = ""
)
//GraphVariable represents a variable before it is parsed and written into a query.
//TODO: Minimize interface{} here.
type GraphVariable struct {
	Type  VarType
	Value interface{}
}

//Function represents a GraphQL+- function. It writes into the query
//the function type, checks the type as well as the list of arguments.
//It uses GraphQL variables to minimize risk of any type of injection.
type Function struct {
	Type      functionType
	Variables []GraphVariable
	mapValues []string
	Order     []Ordering
}

func MakeFunction(ft functionType) *Function {
	return &Function{Type: ft}
}

func (f *Function) AddOrdering(t OrderType, pred string) *Function {
	f.Order = append(f.Order, Ordering{Type: t, Predicate: Predicate(pred)})
	return f
}

func (f *Function) AddValue(v interface{}, t VarType) *Function {
	vv := GraphVariable{t, v}
	f.Variables = append(f.Variables, vv)
	return f
}

func (f *Function) AddValues(t VarType, v ...interface{}) *Function {
	for k := range v {
		v2 := GraphVariable{t, v[k]}
		f.Variables = append(f.Variables, v2)
	}
	return f
}

func (f *Function) AddPred(name string) *Function {
	vv := GraphVariable{TypePred, name}
	f.Variables = append(f.Variables, vv)
	return f
}

func (f *Function) AddPredValue(name string, v interface{}, t VarType) *Function {
	v1 := GraphVariable{TypePred, name}
	v2 := GraphVariable{t, v}
	f.Variables = append(f.Variables, v1, v2)
	return f
}

func (f *Function) AddMatchValues(name string, v string, count int) *Function {
	v1 := GraphVariable{TypePred, name}
	v2 := GraphVariable{TypeStr, v}
	v3 := GraphVariable{TypeDefault, count}
	f.Variables = append(f.Variables, v1, v2, v3)
	return f
}

func (f *Function) AddPredMultiple(name string, t VarType, v ...interface{}) *Function {
	v1 := GraphVariable{TypePred, name}
	f.Variables = append(f.Variables, v1)
	for k := range v {
		v2 := GraphVariable{t, v[k]}
		f.Variables = append(f.Variables, v2)
	}
	return f
}

func (f *Function) mapVariables(q *GeneratedQuery) {
	f.mapValues = make([]string, 0, 2)
	var slice []string
	for _, v := range f.Variables {
		//Handle the two special cases that do not need variable mapping.
		if v.Type == TypePred {
			slice = append(slice, "<"+v.Value.(string)+">")
			continue
		}
		if v.Type == TypeUid {
			if len(v.Value.(string)) > 12 {
				panic("invalid UID, this could be an SQL injection.")
			}
			slice = append(slice, "\""+v.Value.(string)+"\"")
			continue
		}
		/*if v.Type == TypeDefault {
			slice = append(slice, fmt.Sprintf("%v", v.Value))
			continue
		}*/
		//Build the variable using the integer from the query.
		key := q.registerVariable(v.Type, fmt.Sprintf("%v", v.Value))
		slice = append(slice, key)
	}
	f.mapValues = slice
}

func (f *Function) create(q *GeneratedQuery, parent string, sb *bytes.Buffer) {
	//No nil checks etc. Should be done before.
	//Map the variables to their proper value.
	f.mapVariables(q)
	//Write the default values.
	sb.WriteString(strings.ToLower(string(f.Type)))
	sb.WriteString(tokenLP)
	sb.WriteString(f.buildVarString())
	sb.WriteString(tokenRP)
	for k := range f.Order {
		sb.WriteString(tokenComma)
		s := f.Order[k].String()
		sb.WriteString(s)
	}
}

func (f *Function) buildVarString() string {
	var sb bytes.Buffer
	for k, v := range f.mapValues {
		sb.WriteString(v)
		if k != len(f.Variables)-1 {
			sb.WriteByte(',')
		}
	}
	return sb.String()
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

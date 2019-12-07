package humus

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

var pool sync.Pool

func init() {
	pool.New = func() interface{} {
		return new(GeneratedQuery)
	}
}

/*
	UID represents the primary UID class used in communication with DGraph.
	This is used in code generation.
*/
type UID string

func (u UID) Int() int64 {
	if len(u) < 2 {
		return -1
	}
	val, err := strconv.ParseInt(string(u[2:]), 16, 64)
	if err != nil {
		return -1
	}
	return val
}

func (u UID) IntString() string {
	val, err := strconv.ParseInt(string(u), 16, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatInt(val, 10)
}

func StringFromInt(id int64) string {
	return "0x" + strconv.FormatInt(id, 16)
}

func UidFromInt(id int64) UID {
	return UID(StringFromInt(id))
}

const (
	// syntax tokens
	tokenLB     = "{" // Left Brace
	tokenRB     = "}" // Right Brace
	tokenLP     = "(" // Left parenthesis
	tokenRP     = ")" // Right parenthesis
	tokenColumn = ":"
	tokenComma  = ","
	tokenSpace  = " "
	tokenFilter = "@filter"
)

//TODO: These have not been tested yet. Only GeneratedQuery.
//TODO: Should this allow arbitrary Query interfaces and type switch?
type Queries struct {
	q          []*GeneratedQuery
	varCounter func() int
	currentVar int
	vars       map[string]string
}

//Satisfy the Query interface.
func (q *Queries) Process() (string, error) {
	return q.create()
}

func (q *Queries) NewQuery(f Fields) *GeneratedQuery {
	newq := &GeneratedQuery{
		modifiers: make(map[Predicate]modifierList),
		fields:    f,
		varMap: q.vars,
	}
	newq.varFunc = q.varCounter
	q.q = append(q.q, newq)
	return newq
}

//create the byte representation.
func (q *Queries) create() (string, error) {
	var final strings.Builder
	final.Grow(512)
	//The query variable information.
	final.WriteString("query t")
	//The global variable counter. It exists in a closure, it's just easy.
	final.WriteString("(")
	for k, qu := range q.q {
		qu.mapVariables(qu)
		str := qu.variables()
		if str == "" {
			continue
		}
		//TODO: Make it more like a strings.Join to avoid all these error-prone additions.
		if len(q.q) > 1 && k > 0 {
			final.WriteByte(',')
		}
		final.WriteString(str)
	}
	final.WriteByte(')')
	for k, qu := range q.q {
		final.WriteByte('{')
		qu.index = k + 1
		_, err := qu.create(&final)
		if err != nil {
			return "", err
		}
		final.WriteByte('}')
	}
	return final.String(), nil
}

func (q *Queries) Vars() map[string]string {
	return q.vars
}

//The Field maps include a path for the predicate.
//Root is "", all sub are /Predicate1/Predicate2...
//It is quite a big allocation.
type GeneratedQuery struct {
	//The root function.
	//Since all queries have a graph function this is an embedded struct.
	//(It is embedded for convenient access as well as unnecessary pointer traversal).
	function
	//If it is a var query.
	v bool
	//Top level filter.
	//filter *Filter
	//List of modifiers, i.e. order, pagination etc.
	modifiers map[Predicate]modifierList
	//Builder for variables.
	varBuilder strings.Builder
	//Map for dealing with GraphQL variables.
	varMap map[string]string
	//function for getting next query value in multi-query.
	varFunc func() int
	//The overall language for this query.
	language Language
	//Whether to allow untagged.
	strictLanguage bool
	//Which directives to apply on this query.
	directives []Directive
	//The list of fields
	fields Fields
	//Current GraphQL variable.
	varCounter int
	//For multiple queries.
	index int
}

/*
func (q *GeneratedQuery) Var(v bool) *GeneratedQuery {
	q.v = v
	return q
}
*/
//Facets sets @facets for the edge specified by path.
func (q *GeneratedQuery) Facets(path Predicate) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], facet{})
	return q
}
//NewQuery returns a new singular generation query for use
//in building a single query.
func NewQuery(f Fields) *GeneratedQuery {
	return &GeneratedQuery{
		varMap:    make(map[string]string),
		modifiers: make(map[Predicate]modifierList),
		fields:    f,
	}
}
//NewQueries returns a QueryList used for building
//multiple queries at once.
func NewQueries() *Queries {
	qu := new(Queries)
	qu.q = make([]*GeneratedQuery, 0, 2)
	qu.currentVar = -1
	qu.varCounter = func() int {
		qu.currentVar++
		return qu.currentVar
	}
	qu.vars = make(map[string]string)
	return qu
}

func (q *GeneratedQuery) Process() (string, error) {
	return q.create(nil)
}

//Order adds an ordering of type t at the given object path to the predicate pred.
//Requires is that the path points to an object predicate where pred is a predicate in the given object.
func (q *GeneratedQuery) Order(t OrderType, path Predicate, pred Predicate) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], Ordering{Type: t, Predicate: pred})
	return q
}

type MutationType string

const (
	MutateDelete MutationType = "delete"
	MutateSet    MutationType = "set"
)

func (q *GeneratedQuery) create(sb *strings.Builder) (string, error) {
	//t := time.Now()
	//sb == nil implies this is a solo query. This means we need to map the GraphQL
	//variables beforehand as it is otherwise calculated in the Queries calculation in Queries.create()
	if sb == nil {
		sb = new(strings.Builder)
		q.mapVariables(q)
		sb.Grow(512)
	}
	if err := q.function.check(q); err != nil {
		return "", err
	}
	//Single query.
	if q.varFunc == nil {
		vars := q.variables()
		if vars == "" {
			sb.WriteString("query{")
		} else {
			sb.WriteString("query t(")
			sb.WriteString(vars)
			sb.WriteString("){")
		}
	}
	//Write query header.
	sb.WriteByte('q')
	if q.index != 0 {
		writeInt(int64(q.index), sb)
	}
	sb.WriteString(tokenLP + "func" + tokenColumn + tokenSpace)
	err := q.function.create(q, sb)
	if err != nil {
		return "", err
	}
	//Top level modifiers.
	val, ok := q.modifiers[""]
	if ok {
		//Two passes. Before and after parenthesis. That's just how it be.
		sort.Sort(val)
		for _, v := range val {
			if v.priority() > modifierFilter && v.canApply(modifierFunction) {
				sb.WriteByte(',')
				err := v.apply(q, 0, modifierFunction, sb)
				if err != nil {
					return "", err
				}
			}
		}
		//TODO: Check this through.
		sb.WriteByte(')')
		for _, v := range val {
			if v.priority() == modifierFilter && v.canApply(modifierFunction) {
				err := v.apply(q, 0, modifierFunction, sb)
				if err != nil {
					return "", err
				}
			}
		}
	} else {
		sb.WriteByte(')')
	}
	for _, v := range q.directives {
		sb.WriteByte('@')
		sb.WriteString(string(v))
	}
	sb.WriteByte('{')
	var parentBuf = make(unsafeSlice, 0, 64)
	for _, field := range q.fields.Get() {
		if len(field.Name) > 64 {
			//This code should pretty much never execute as a predicate is rarely this large.
			parentBuf = make([]byte, 2*len(field.Name))
		}
		parentBuf = parentBuf[:len(field.Name)]
		copy(parentBuf, field.Name)
		//parentBuf = append(parentBuf, field.Name...)
		err := field.create(q, parentBuf, sb)
		if err != nil {
			return "", err
		}
		sb.WriteByte(' ')
	}
	for _, v := range val {
		if v.priority() <= modifierAggregate && v.canApply(modifierField) {
			err := v.apply(q, 0, modifierFunction, sb)
			if err != nil {
				return "", err
			}
		} else {
			break
		}
	}
	//Add default uid to top level field and close query.
	sb.WriteString(" uid" + tokenRB)
	if q.varFunc == nil {
		sb.WriteByte('}')
	}
	return sb.String(), nil
}

func (q *GeneratedQuery) Vars() map[string]string {
	return q.varMap
}
//Directive adds a top level directive.
func (q *GeneratedQuery) Directive(dir Directive) *GeneratedQuery {
	for _, v := range q.directives {
		if v == dir {
			return q
		}
	}
	q.directives = append(q.directives, dir)
	return q
}

//Count adds a count to a predicate specified by the path.
func (q *GeneratedQuery) Count(t CountType, path Predicate, value int) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], Pagination{Type: t, Value: value})
	return q
}

//Agg adds aggregation on a value variable.
//Note that path here is the root level of the aggregation such that
//empty path corresponds to top level aggregation.
//The variable represents which value to aggregate on and alias
//is the alias for this aggregation.
func (q *GeneratedQuery) Agg(typ AggregateType, path Predicate, variable string, alias string) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], AggregateValues{Type: typ, Alias: alias, Variable: variable})
	return q
}
//GroupBy sets groupby on the path.
//TODO: Make this work properly.
func (q *GeneratedQuery) GroupBy(path Predicate, value Predicate) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], groupBy(value))
	return q
}

//Filter adds a subfilter to the edge specified by path.
//If path is "Node.edge" then the first edge from Node will have a filter
//applied alongside the "Node.edge" predicate.
func (q *GeneratedQuery) Filter(f *Filter, path Predicate) *GeneratedQuery {
	f.mapVariables(q)
	q.modifiers[path] = append(q.modifiers[path], f)
	return q
}

//Language sets the language for the query to apply to all fields.
//If strict do not allow untagged language.
func (q *GeneratedQuery) Language(l Language, strict bool) *GeneratedQuery {
	q.language = l
	q.strictLanguage = strict
	return q
}

// Returns all the query variables for this query in create form.
//single for single query
func (q *GeneratedQuery) variables() string {
	return q.varBuilder.String()
}

//Variable adds a value variable of the form name as value.
//Value can be anything from a math expression to a count.
//Example: if name is test and value is math(p+q) then
//result is a variable 'test as math(p+q)' at the path given by path.
//isAlias simply means that rather than a value variable it is
//test : math(p+q)
func (q *GeneratedQuery) Variable(name string, path Predicate, value string, isAlias bool) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], variable{
		name:  name,
		value: value,
		alias: isAlias,
	})
	return q
}

func (q *GeneratedQuery) registerVariable(typ varType, value string) string {
	if q.varBuilder.Len() != 0 {
		q.varBuilder.WriteByte(',')
	} else {
		q.varBuilder.Grow(32)
	}
	var val int
	if q.varFunc != nil {
		val = q.varFunc()
	} else {
		val = q.varCounter
		q.varCounter++
	}
	var buf [9]byte
	buf[0] = '$'
	b := strconv.AppendInt(buf[:1], int64(val), 10)
	key := string(b)
	q.varBuilder.WriteString(key)
	q.varBuilder.WriteByte(':')
	q.varBuilder.WriteString(string(typ))
	q.varMap[key] = value
	return key
}

//Static create a static query from the generated version.
//Since this is performed at init, panic if the query
//creation does not work. V
func (q *GeneratedQuery) Static() StaticQuery {
	str, err := q.create(nil)
	if err != nil {
		panic(err)
	}
	return StaticQuery{
		Query: str,
		vars:  nil,
	}
}

//Function sets the function type for this function. It is used alongside
//variables. Variables are automatically mapped to GraphQL variables as a way
//of avoiding SQL injections.
func (q *GeneratedQuery) Function(ft FunctionType) *GeneratedQuery {
	q.function = function{Type: ft, Variables: make([]graphVariable, 0, 4)}
	return q
}

//Pred sets a predicate variable, for a has function.
func (q *GeneratedQuery) Pred(pred Predicate) *GeneratedQuery {
	q.function.pred(pred)
	return q
}

//PredValue sets a predicate alongside a value, useful for eq.
func (q *GeneratedQuery) PredValue(pred Predicate, value interface{}) *GeneratedQuery {
	q.function.predValue(pred, value)
	return q
}
//PredValues sets a predicate value alongside an array of values.
//E.g. the equals function with multiple values.
func (q *GeneratedQuery) PredValues(pred Predicate, value ...interface{}) *GeneratedQuery {
	q.function.predMultiple(pred, value)
	return q
}
//Value sets a single value, i.e. for has.
func (q *GeneratedQuery) Value(v interface{}) *GeneratedQuery {
	q.function.value(v)
	return q
}
//Values sets a multiple values of any type.
func (q *GeneratedQuery) Values(v ...interface{}) *GeneratedQuery {
	q.function.values(v)
	return q
}

package mulbase

import (
	"bytes"
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
	Queries []*GeneratedQuery
}

//Satisfy the Query interface.
func (q *Queries) Process(SchemaList) ([]byte, map[string]string, error) {
	return q.create()
}


func (q *Queries) Append(qu *GeneratedQuery) *Queries {
	q.Queries = append(q.Queries, qu)
	return q
}

//create the byte representation.
func (q *Queries) create() ([]byte, map[string]string, error) {
	var queryStr, final bytes.Buffer
	//The query variable information.
	final.WriteString("query t")
	//The global variable counter. It exists in a closure, it's just easy.
	var varCounter = -1
	//TODO: Better to not use a closure as to avoid heap allocating an int..
	varFunc := func() int {
		varCounter++
		return varCounter
	}
	var output = make(map[string]string)
	for k, qu := range q.Queries {
		qu.index = k + 1
		qu.varMap = output
		qu.varFunc = varFunc
		str, err := qu.create()
		if err != nil {
			return nil, nil, err
		}
		queryStr.WriteString(str)
	}
	final.WriteString("(")
	for k, qu := range q.Queries {
		str := qu.Variables()
		if str == "" {
			continue
		}
		//TODO: Make it more like a strings.Join to avoid all these error-prone additions.
		if len(q.Queries) > 1 && k > 0 {
			final.WriteByte(',')
		}
		final.WriteString(str)
	}
	final.WriteString("){")
	final.WriteString(queryStr.String())
	final.WriteString("}")
	return final.Bytes(), output, nil
}

//The Field maps include a path for the predicate.
//Root is "", all sub are /Predicate1/Predicate2...
//It is quite a big allocation.
//TODO: All maps kept separate? They might not be used that often.
type GeneratedQuery struct {
	//The root function.
	Function *Function
	//Top level filter.
	Filter *Filter
	//All sub parts of the query.
	//FieldFunctions map[Predicate]Function
	//FieldOrderings map[Predicate][]Ordering
	//FieldCount     map[Predicate][]Pagination
	//FieldAggregate map[Predicate]AggregateValues
	//FieldFilters   map[Predicate]Filter
	modifiers  map[Predicate]modifierList
	varBuilder strings.Builder
	varMap     map[string]string
	varFunc    func() int
	//The overall language for this query.
	Language Language
	strictLanguage bool
	//Which directives to apply on this query.
	Directives []Directive
	Fields     Fields
	varCounter int
	schema     SchemaList
	//For multiple queries.
	index int
}

func (q *GeneratedQuery) SetFields(f Fields) *GeneratedQuery {
	q.Fields = f
	return q
}

func (q *GeneratedQuery) Facets(path Predicate) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], facet{})
	return q
}

func NewQuery() *GeneratedQuery {
	return &GeneratedQuery{
		varMap:    make(map[string]string),
		modifiers: make(map[Predicate]modifierList),
	}
}

func (q *GeneratedQuery) Process(sch SchemaList) (string, error) {
	q.schema = sch
	return q.create()
}

func (q *GeneratedQuery) AddOrdering(t OrderType, path Predicate, pred Predicate) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], Ordering{Type: t, Predicate: pred})
	return q
}

type MutationType string

const (
	MutateDelete MutationType = "delete"
	MutateSet    MutationType = "set"
)

func (q *GeneratedQuery) create() (string, error) {
	//t := time.Now()
	var sb strings.Builder
	//The size of default buffer.
	const size = 256
	sb.Grow(size)
	//Write query header.
	sb.WriteString("{q")
	if q.index != 0 {
		sb.WriteString(strconv.Itoa(q.index))
	}
	sb.WriteString(tokenLP + "func" + tokenColumn + tokenSpace)
	err := q.Function.create(q, &sb)
	if err != nil {
		return "", err
	}
	sb.WriteString(tokenRP)
	//optional filter
	if q.Filter != nil {
		err := q.Filter.create(q, &sb)
		if err != nil {
			return "", err
		}
	}
	for _, v := range q.Directives {
		sb.WriteByte('@')
		sb.WriteString(string(v))
	}
	sb.WriteString(tokenLB)
	var parentBuf = make(unsafeSlice,0,64)
	for i, field := range q.Fields.Get() {
		if i != 0 {
			sb.WriteString(tokenSpace)
		}
		parentBuf = append(parentBuf, field.Name...)
		err := field.create(q, parentBuf, &sb)
		if err != nil {
			return "", err
		}
		parentBuf = parentBuf[:0]
	}
	//Add default uid to top level field and close query.
	sb.WriteString(" uid" + tokenRB + tokenRB)
	//TODO: Write variable header and create the var map.
	//TODO: This is a way to see if it is single. This is not the best!
	//TODO: Is there anyway to preallocate variables? I don't think so. That would imply
	//traversing the tree once beforehand and that is unnecessary.
	vars := q.Variables()
	var result strings.Builder
	if q.varFunc == nil {
		result.Grow(len(vars)+sb.Len() + len("query t(") + 1)
	} else {
		result.Grow(sb.Len())
	}
	//Copy the query into result.
	if q.varFunc == nil {
		result.WriteString("query t(")
		result.WriteString(vars)
		result.WriteString(")")
	}
	result.WriteString(sb.String())
	//fmt.Println(fmt.Sprintf("Creating query took %v", time.Now().Sub(t)))
	buf := result.String()
	return buf, nil
}

func (q *GeneratedQuery) Vars() map[string]string {
	return q.varMap
}

func (q *GeneratedQuery) AddDirective(dir Directive) *GeneratedQuery {
	for _, v := range q.Directives {
		if v == dir {
			return q
		}
	}
	q.Directives = append(q.Directives, dir)
	return q
}

//Adds a count to a predicate.
func (q *GeneratedQuery) AddSubCount(t PaginationType, path Predicate, value int) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], Pagination{Type: t, Value: value})
	return q
}
//AddAggregation adds aggregation on a value variable.
//Note that path here is the root level of the aggregation such that
//empty path corresponds to top level aggregation.
func (q *GeneratedQuery) AddAggregation(typ AggregateType, path Predicate, alias string) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], AggregateValues{Type:typ, Alias: alias})
	return q
}

//Adds a subfilter to a predicate.
func (q *GeneratedQuery) AddSubFilter(f *Function, path Predicate) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], MakeFilter(f))
	return q
}

//SetLanguage sets the language for the query to apply to all fields.
//If strict do not allow untagged language.
func (q *GeneratedQuery) SetLanguage(l Language, strict bool) *GeneratedQuery {
	q.Language = l
	return q
}

// Returns all the query variables for this query in create form.
//single for single query
func (q *GeneratedQuery) Variables() string {
	if q.varBuilder.Len() == 0 {
		return ""
	}
	return q.varBuilder.String()
}

//The alias is to avoid count(predicate) as name.
func (q *GeneratedQuery) SetSubAggregate(path Predicate, alias string, aggregate AggregateType) *GeneratedQuery {
	q.modifiers[path] = append(q.modifiers[path], AggregateValues{aggregate, alias})
	return q
}

//It does not build it concurrently so just increment a counter.
func (q *GeneratedQuery) registerVariable(typ VarType, value string) string {
	if q.varBuilder.Len() != 0 {
		q.varBuilder.WriteByte(',')
	}
	var val int
	if q.varFunc != nil {
		val = q.varFunc()
	} else {
		val = q.varCounter
		q.varCounter++
	}
	var key = "$" + strconv.Itoa(val)
	q.varBuilder.WriteString(key)
	q.varBuilder.WriteByte(':')
	q.varBuilder.WriteString(string(typ))
	q.varMap[key] = value
	return key
}

func (q *GeneratedQuery) SetFunction(function *Function) *GeneratedQuery {
	q.Function = function
	return q
}

//TODO: Multiple filters.
func (q *GeneratedQuery) SetFilter(filter *Filter) *GeneratedQuery {
	q.Filter = filter
	return q
}
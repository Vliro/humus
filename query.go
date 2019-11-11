package mulbase

import (
	"bytes"
	"strconv"
	"strings"
)

/**
	UID represents the primary UID class used in communication with DGraph.
	This is used in code generation.
 */
type UID string

func (u UID) Int() int64 {
	val, err := strconv.ParseInt(string(u), 16, 64)
	if err != nil {
		return -1
	}
	return val
}

type H map[string]interface{}

type DNode interface {
	UID() UID
	SetUID(uid string)
	SetType()
	Fields() FieldList
	//Serializes all the scalar values that are not hidden.
	Values() map[string]interface{}
}

const (
	// syntax tokens
	tokenLB     = "{" // Left Brace
	tokenRB     = "}" // Right Brace
	tokenLP     = "(" // Left Parenthesis
	tokenRP     = ")" // Right Parenthesis
	tokenColumn = ":"
	tokenComma  = ","
	tokenSpace  = " "
	tokenFilter = "@filter"
)


type Queries struct {
	Queries []*GeneratedQuery
}

func (q *Queries) Process(schemaList) ([]byte, map[string]string, error) {
	return q.create()
}

func (q *Queries) Type() QueryType {
	return QueryRegular
}

func (q *Queries) Append(qu *GeneratedQuery) *Queries {
	q.Queries = append(q.Queries, qu)
	return q
}

func mapMerge(m1,m2 map[string]string) {
	for k,v := range m2 {
		m1[k] = v
	}
}

func (q *Queries) create() ([]byte,map[string]string, error) {
	queryStr := bytes.Buffer{}
	final := bytes.Buffer{}
	//The query variable information.
	final.WriteString("query t")
	//The global variable counter.
	var varCounter int
	//closure
	varFunc := func() int {
		varCounter++
		return varCounter - 1
	}
	var e error
	var output = make(map[string]string)
	for k, qu := range q.Queries {
		e = qu.check()
		if e != nil {
			return nil, nil, e
		}
		qu.index = k + 1
		qu.VarMap = output
		qu.VarFunc = varFunc
		str, _, err := qu.create()
		if err != nil {
			return nil, nil, err
		}
		queryStr.Write(str)
	}
	final.WriteString("(")
	for k, qu := range q.Queries {
		str := qu.Variables(false)
		if str == "" {
			continue
		}
		if len(q.Queries) > 1 && k > 0 {
			final.WriteByte(',')
		}
		final.WriteString(str)
	}
	final.WriteString(")")
	final.WriteString("{")
	final.WriteString(queryStr.String())
	final.WriteString("}")
	return final.Bytes(), output, nil
}

type VarObject struct {
	val     string
	varType VarType
}

//An aggregate value i.e. sum as well as what alias to name it as.
type aggregateValues struct {
	Type  AggregateType
	Alias string
}

//The Field maps include a path for the predicate.
//Root is "", all sub are /Predicate1/Predicate2...
type GeneratedQuery struct {
	//The root function.
	Function       *Function
	//All sub parts of the query.
	FieldFunctions map[string]Function
	FieldOrderings map[string][]Ordering
	FieldCount     map[string][]Count
	FieldAggregate map[string]aggregateValues
	FieldFilters   map[string]Filter
	varBuilder 	   strings.Builder
	VarMap         map[string]string
	VarFunc        func() int
	Filter         *Filter
	Language       Language
	//Which directives to apply on this query.
	Directives     []Directive
	Deserialize    bool
	Fields         []Field
	varCounter int
	schema schemaList
	//For multiple queries.
	index int
}

func (q *GeneratedQuery) SetFields(f []Field) *GeneratedQuery {
	q.Fields = f
	return q
}

func NewQuery() *GeneratedQuery {
	return &GeneratedQuery{
		VarMap: make(map[string]string),
	}
}

func (q *GeneratedQuery) Process(sch schemaList) ([]byte, map[string]string, error) {
	q.schema = sch
	return q.create()
}

func (q *GeneratedQuery) Type() QueryType {
	return QueryRegular
}

func (q *GeneratedQuery) SetSubOrdering(t OrderType, path string, pred string) *GeneratedQuery {
	if q.FieldOrderings == nil {
		q.FieldOrderings = make(map[string][]Ordering)
	}
	val := q.FieldOrderings[path]
	val = append(val, Ordering{Type: t, Predicate: Predicate(pred)})
	q.FieldOrderings[path] = val
	return q
}

type MutationType string

const (
	MutateDelete MutationType = "delete"
	MutateSet    MutationType = "set"
)

type Mutation struct {
	Type   MutationType
	Object interface{}
}

func (m *Mutation) AddValue(val interface{}) *Mutation {
	m.Object = val
	return m
}

func (q *GeneratedQuery) create() ([]byte, map[string]string, error) {
	if err := q.check(); err != nil {
		return nil, nil, err
	}
	var sb = &bytes.Buffer{}
	//Write query header.
	sb.WriteString("{q")
	if q.index != 0 {
		sb.WriteString(strconv.Itoa(q.index))
	}
	sb.WriteString(tokenLP)
	sb.WriteString("func")
	sb.WriteString(tokenColumn)
	sb.WriteString(tokenSpace)
	q.Function.create(q, "", sb)
	sb.WriteString(tokenRP)
	//optional filter
	q.Filter.create(q, "", sb)
	for _, v := range q.Directives {
		sb.WriteString("@" + string(v))
	}
	sb.WriteString(tokenLB)
	for i, field := range q.Fields {
		if i != 0 {
			sb.WriteString(tokenSpace)
		}
		err := field.String(q, field.Name, sb)
		if err != nil {
			return nil, nil, err
		}
	}
	sb.WriteString(tokenSpace + "uid" + tokenSpace + tokenRB + tokenRB)
	//TODO: Write variable header and create the var map.
	var varString = q.Variables(true)
	var result = make([]byte, len(varString) + len(sb.Bytes()))
	//Copy the query into result.
	copy(result, varString)
	copy(result[len(varString):], sb.Bytes())

	return result, q.VarMap, nil
}

func (q *GeneratedQuery) check() error {
	if err := q.Function.check(q); err != nil {
		return err
	}
	for _, v := range q.FieldFunctions {
		if err := v.check(q); err != nil {
			return err
		}
	}
	for _, v := range q.FieldFilters {
		if err := v.check(q); err != nil {
			return err
		}
	}
	if err := q.Function.check(q); err != nil {
		return err
	}
	// check query
	return nil
}

//Adds a count to a predicate.
func (q *GeneratedQuery) AddSubCount(t CountType, path string, value int) *GeneratedQuery {
	c := Count{Type: t, Value: value}
	if q.FieldCount == nil {
		q.FieldCount = make(map[string][]Count)
	}
	val := q.FieldCount[path]
	//can append to nil slice:)
	q.FieldCount[path] = append(val, c)
	return q
}

//Adds a subfilter to a predicate.
func (q *GeneratedQuery) AddSubFilter(f *Function, path string, logical ...string) *GeneratedQuery {
	if q.FieldFilters == nil {
		q.FieldFilters = make(map[string]Filter)
	}
	val, ok := q.FieldFilters[path]
	if ok {
		//All connected functions must have logical
		//TODO: if already exists
		return q
	} else {
		val = Filter{}
		val.Function = f
		q.FieldFilters[path] = val
		return q
	}
}

//SetLanguage sets the language for the query to apply to all fields.
//Default english.
func (q *GeneratedQuery) SetLanguage(l Language) *GeneratedQuery {
	q.Language = l
	return q
}

// Returns all the query variables for this query in create form.
//single for single query
func (q *GeneratedQuery) Variables(single bool) string {
	/*sb := bytes.Buffer{}
	i := 0
	sb.WriteString("query test(")
	for k, v := range q.VarMap {

		if v.varType == TypeStr || v.varType == TypeInt {
			sb.WriteString(k)
			sb.WriteString(": ")
			sb.WriteString(string(v.varType))
		}
		if i != len(q.VarMap)-1 {
			sb.WriteByte(',')
		}
		i++
	}
	sb.WriteByte(')')*/
	if single {
		return "query test(" + q.varBuilder.String() + ")"
	}
	return q.varBuilder.String()
}

//The alias is to avoid count(predicate) as name.
func (q *GeneratedQuery) SetSubAggregate(path string, alias string, aggregate AggregateType) *GeneratedQuery {
	if q.FieldAggregate == nil {
		q.FieldAggregate = make(map[string]aggregateValues)
	}
	q.FieldAggregate[path] = aggregateValues{aggregate, alias}
	return q
}

//It does not build it concurrently so just increment a counter.
func (q *GeneratedQuery) registerVariable(typ VarType, value string) string {
	if q.varBuilder.Len() != 0 {
		q.varBuilder.WriteByte(',')
	}
	var val int
	if q.VarFunc != nil {
		val = q.VarFunc()
	} else {
		val = q.varCounter
		q.varCounter++
	}
	var key = "$"+strconv.Itoa(val)
	q.varBuilder.WriteString(key+":"+string(typ))
	q.VarMap[key] = value
	return key
}

// JSON returns a json create with "query" field.
// shouldVar tells if it should print out the GraphQL variable create.
/*
func (q *GeneratedQuery) JSON(shouldVar bool) ([]byte, error) {
	if q.VarMap == nil {
		q.VarMap = make(map[create]VarObject)
	}
	if q.Type == TypeMutate {
		if len(q.Mutations) == 0 {
			return nil, fmt.Errorf("no mutations")
		} else {
			m := q.Mutations[0]
			q.getMutateType = m.Type
		}
	}
	s, err := q.Create()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if q.Type == TypeQuery && shouldVar {
		var buf bytes.Buffer
		buf.WriteString(q.createVariableString())
		buf.Write(s)
		return buf.Bytes(), err
	}
	return s, err
}*/

func (q *GeneratedQuery) SetFunction(function *Function) *GeneratedQuery {
	q.Function = function
	return q
}

//TODO: Multiple filters.
func (q *GeneratedQuery) SetFilter(filter *Filter) *GeneratedQuery {
	q.Filter = filter
	return q
}

// ShouldDeserialize defines if either an interface{} is returned or it is deserialized to the proper object.
func (q *GeneratedQuery) ShouldDeserialize(b bool) *GeneratedQuery {
	q.Deserialize = b
	return q
}

/*
func DeleteDNode(d DNode, ctx context.Context, sync bool, txn *TxnQuery) (map[create]create, error) {
	m := Mutation{}
	m.Object = d.DeleteUIDS()
	m.Type = MutateDelete
	return MutateMany(ctx, sync, nil, m)
}

func CreateDNode(js DNode, sync bool, txn *TxnQuery) (map[create]create, error) {
	js.SetType()
	return saveNodeInternal(js, sync, txn)
}

func saveInterface(ctx context.Context, txn *TxnQuery, v interface{}) (map[create]create, error) {
	m := Mutation{}
	m.Type = MutateSet
	m.Object = v
	return MutateMany(ctx, true, txn, m)
}

func saveInterfaceBuffer(txn *TxnQuery, v interface{}) {
	m := Mutation{}
	m.Object = v
	m.Type = MutateSet
	MutateMany(context.Background(), false, txn, m)
}

func deleteInterfaceBuffer(sync bool, txn *TxnQuery, val interface{}) {
	m := Mutation{}
	m.Object = val
	m.Type = MutateDelete
	MutateMany(context.Background(), false, txn, m)
}

func saveNodeInternal(js DNode, sync bool, txn *TxnQuery) (map[create]create, error) {
	m := Mutation{}
	m.Object = js.GetAllInfo(true)
	m.Type = MutateSet
	return MutateMany(context.Background(), sync, nil, m)
}

func SaveMultiplePreds(uid create, pred []create, ctx context.Context, sync bool, txn *TxnQuery, vals ...interface{}) error {
	if len(pred) != len(vals) {
		return errors.New("saveMultiple: invalid pred to vals len")
	}
	m := make(H)
	m["uid"] = uid
	for k, v := range pred {
		m[v] = vals[k]
	}
	mut := Mutation{Type: MutateSet, Object: m}
	_, err := MutateMany(ctx, sync, txn, mut)
	return err
}

func SaveDNodes(sync bool, txn *TxnQuery, js ...DNode) (map[create]create, error) {
	var m = make([]Mutation, len(js))
	for k, v := range js {
		v.SetType()
		m[k] = Mutation{
			Type:   MutateSet,
			Object: v.GetAllInfo(true),
		}
	}
	if txn != nil {
		txn.AddManyMutations(m...)
		return txn.ExecuteLatest(nil, nil)
	}
	return MutateMany(nil, sync, txn, m...)
}

//Returns UID for the "root" node mutation, they are of the form "blank-I", i order of mutation.
func GetRootUID(ma map[create]create) create {
	v, _ := ma["blank-0"]
	return v
}

func (q *GeneratedQuery) AddDNode(d DNode, new bool, typ MutationType) *GeneratedQuery {
	if new || d.UID() == "" {
		d.SetType()
	}
	var m Mutation
	if typ == MutateSet {
		m.Type = MutateSet
		m.Object = d.GetAllInfo(true)
	} else if typ == MutateDelete {
		m.Type = MutateDelete
		m.Object = d.DeleteUIDS()
	}
	q.Mutations = append(q.Mutations, m)
	return q
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
*/
/*
Returns a Queries object which can be used for multiple queries.
*/
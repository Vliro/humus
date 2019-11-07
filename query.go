package ngraph

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/dgraph-io/dgo"

	"github.com/pkg/errors"
)

/**
	UID represents the primary UID class used in communication with DGraph.
 */
type UID string

type H map[string]interface{}

type DNode interface {
	UID() string
	SetUID(uid string)
	SetType()
	//Use this to recursively get uids to properly perform deletion.
	DeleteUIDS() map[string]interface{}
}

type BaseNode struct {
	Uid  string   `json:"uid"`
	Type []string `json:"dgraph.type,omitempty"`
}

func (b *BaseNode) SetUID(uid string) {
	b.Uid = uid
}

func (b *BaseNode) AddType(typ string) {
	if b.Uid != "" {
		return
	}
	b.Type = append(b.Type, typ)
}

func (b *BaseNode) GetAllInfo() map[string]interface{} {
	return makeUIDMap(b.Uid)
}

func (b *BaseNode) GetRelativeInfo() map[string]interface{} {
	return makeUIDMap(b.Uid)
}

func (b *BaseNode) DeleteUIDS() map[string]interface{} {
	return makeUIDMap(b.Uid)
}

func (b *BaseNode) UID() string {
	return b.Uid
}

func (b *BaseNode) SetType() {
}

type operationType string

// 2 types of operation.
const (
	TypeQuery  operationType = "query"
	TypeMutate operationType = "mutation"
)

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

var (
//json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type Queries struct {
	Queries []*Query
}

func (q *Queries) Append(qu *Query) *Queries {
	q.Queries = append(q.Queries, qu)
	return q
}

func (q *Queries) Execute(params ...interface{}) error {
	return PerformQueries(q, params)
}

func (q *Queries) Stringify() (string, error) {
	queryStr := bytes.Buffer{}
	final := bytes.Buffer{}
	//The query variable information.
	final.WriteString("query t")
	//The global variable counter.
	var varCounter uint32
	varFunc := func() uint32 {
		varCounter++
		return varCounter - 1
	}
	var e []error
	for k, qu := range q.Queries {
		if qu.Type != TypeQuery {
			return "", errors.Errorf("Queries should only contain TypeQueries")
		}
		qu.Name = "q" + strconv.Itoa(k)
		qu.single = false
		qu.VarFunc = varFunc
		js, err := qu.JSON(false)
		if err != nil {
			e = append(e, err)
		}
		queryStr.Write(js)
	}
	if len(e) > 0 {
		return "", e[0]
	}
	first := true
	for k, qu := range q.Queries {
		str := qu.Variables()
		if str == "" {
			continue
		}
		if len(q.Queries) > 1 && k > 0 {
			final.WriteByte(',')
		}
		if first {
			final.WriteString("(")
			first = !first
		}
		final.WriteString(str)
	}
	if !first {
		final.WriteByte(')')
	}
	final.WriteString("{")
	final.WriteString(queryStr.String())
	final.WriteString("}")
	return final.String(), nil
}

type VarObject struct {
	val     interface{}
	varType VarType
}

type aggregateValues struct {
	Type  AggregateType
	Alias string
}

//The Field maps include a path for the predicate.
//Root is "", all sub are /Predicate1/Predicate2...
type Query struct {
	Type           operationType // The operation type is either query, mutation, or subscription.
	Function       *Function
	JSONMutation   []byte
	FieldFunctions map[string]Function
	FieldOrderings map[string][]Ordering
	FieldCount     map[string][]Count
	FieldAggregate map[string]aggregateValues
	FieldFilters   map[string]Filter
	VarMap         map[string]VarObject
	VarFunc        func() uint32
	Filter         *Filter
	Language       Language
	Mutations      []Mutation
	Directives     []Directive
	Deserialize    bool
	Name           string //= q usually
	Fields         []Field
	varCounter     uint32
	getMutateType  MutationType
	single         bool
	activeTxn      *dgo.Txn
}

func (q *Query) SetSubOrdering(t OrderType, path string, pred string) *Query {
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

func (q *Query) getAllDelete() []Mutation {
	var m []Mutation
	for _, v := range q.Mutations {
		if v.Type == MutateDelete {
			m = append(m, v)
		}
	}
	return m
}
func (q *Query) getAllSet() []Mutation {
	var m []Mutation
	for _, v := range q.Mutations {
		if v.Type == MutateSet {
			m = append(m, v)
		}
	}
	return m
}

func (q *Query) Create() ([]byte, error) {
	if err := q.check(); err != nil {
		return nil, errors.WithStack(err)
	}
	for _, f := range q.Fields {
		if len(f.Name) == 0 {
			return nil, errors.New("error no name")
		}
		if err := f.check(q); err != nil {
			return nil, err
		}
	}
	return q.create().Bytes(), nil
}

type TxnQuery struct {
	Queries []*Query
	txn     *dgo.Txn
	counter int
}

func (t *TxnQuery) ExecuteLatest(ctx context.Context, inp interface{}) (map[string]string, error) {
	if len(t.Queries) < t.counter {
		defer t.txn.Discard(context.Background())
		panic("invalid txnQuery")
	}
	q := t.Queries[t.counter]
	t.counter++
	q.activeTxn = t.txn
	return q.runQuery(ctx, inp)
}

func (t *TxnQuery) AddQuery(q *Query) *TxnQuery {
	t.Queries = append(t.Queries, q)
	return t
}

//Finish up the query and commits the transaction.
func (t *TxnQuery) Finish(discard bool, ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	for k := range t.Queries {
		t.Queries[k] = nil
	}
	if discard {
		return t.txn.Discard(ctx)
	}
	return t.txn.Commit(ctx)
}

func (t *TxnQuery) AddMutation(m Mutation) *TxnQuery {
	q := NewMutate()
	q.Mutations = append(q.Mutations, m)
	return t.AddQuery(q)
}

func (t *TxnQuery) AddManyMutations(m ...Mutation) *TxnQuery {
	q := NewMutate()
	for k := range m {
		q.Mutations = append(q.Mutations, m[k])
	}
	return t.AddQuery(q)
}

//CreateTxn creates a new transaction to use
//for multiple queries/mutations.
//All upcoming mutations/queries on the query acts
//using this transaction instead of the default
//one Txn per operation.
func (q *Query) CreateTxn() *TxnQuery {
	t := TxnQuery{}
	t.Queries = append(t.Queries, q)
	t.txn = getClient().NewTxn()
	return &t
}

func (q *Query) mutateChan(sb *bytes.Buffer) {
	var f []Mutation
	if q.getMutateType == MutateSet {
		f = q.getAllSet()
	} else {
		f = q.getAllDelete()
	}
	for k, v := range f {
		if k != 0 {
			sb.WriteByte(',')
		}
		str := Serialize(v.Object)
		sb.WriteString(str)
	}
}

func (q *Query) setMutateGet(m MutationType) *Query {
	q.getMutateType = m
	return q
}

func (q *Query) create() *bytes.Buffer {
	var sb = &bytes.Buffer{}
	if q.Type == TypeMutate {
		//TODO: this is ugly
		if len(q.Mutations) == 0 {
			return sb
		}
		if len(q.Mutations) > 1 {
			sb.WriteString("[")
		}
		q.mutateChan(sb)
		if len(q.Mutations) > 1 {
			sb.WriteString("]")
		}
		return sb
	}
	if q.single {
		sb.WriteString(tokenLB)
	}
	if q.Name != "" {
		sb.WriteString(tokenSpace)
		sb.WriteString(q.Name)
	} else {
		sb.WriteString("q")
	}
	if q.Type == TypeQuery {
		sb.WriteString(tokenLP)
		sb.WriteString("func")
		sb.WriteString(tokenColumn)
		sb.WriteString(tokenSpace)
		q.Function.string(q, "", sb)
		sb.WriteString(tokenRP)
	}
	//optional filter
	q.Filter.stringChan(q, "", sb)
	for _, v := range q.Directives {
		sb.WriteString("@" + string(v))
	}
	sb.WriteString(tokenLB)
	for i, field := range q.Fields {
		if i != 0 {
			sb.WriteString(tokenSpace)
		}
		field.String(q, field.Name, sb)
	}
	sb.WriteString(tokenSpace)
	sb.WriteString("uid")
	sb.WriteString(tokenSpace)
	sb.WriteString(tokenRB)
	if q.single {
		sb.WriteString(tokenRB)
	}
	return sb
}

func (q *Query) check() error {
	if q.Type == TypeQuery {
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
	}
	// check query
	return nil
}

func (q *Query) checkName() error {
	if q.Name == "" && q.Type == TypeQuery {
		return nil //	return fmt.Errorf("naming error")
	}
	return nil
}

//Adds a count to a predicate.
func (q *Query) AddSubCount(t CountType, path string, value int) *Query {
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
func (q *Query) AddSubFilter(f *Function, path string, logical ...string) *Query {
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
func (q *Query) SetLanguage(l Language) *Query {
	q.Language = l
	return q
}

// MakeQuery constructs a performQuery of the given type and returns a pointer of it.
func makeQuery(Type operationType) *Query {
	return &Query{Type: Type, Deserialize: true, single: true} /*	FieldFunctions: make(map[string]Function), FieldCount: make(map[string][]Count),
		FieldAggregate: make(map[string]aggregateValues),
		FieldFilters:   make(map[string]Filter),
		FieldOrderings: make(map[string][]Ordering)*/
}

func NewQuery() *Query {
	return makeQuery(TypeQuery)
}

func NewMutate() *Query {
	return makeQuery(TypeMutate)
}

func (q *Query) ExecuteFixed(query string, vars map[string]string, inp ...interface{}) error {
	err := executeFix(inp, query, vars)
	return err
}

// Returns all the query variables for this query in string form.
func (q *Query) Variables() string {
	sb := bytes.Buffer{}
	i := 0
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
	return sb.String()
}

//The alias is to avoid count(predicate) as name.
func (q *Query) SetSubAggregate(path string, alias string, aggregate AggregateType) *Query {
	if q.FieldAggregate == nil {
		q.FieldAggregate = make(map[string]aggregateValues)
	}
	q.FieldAggregate[path] = aggregateValues{aggregate, alias}
	return q
}

func (q *Query) createVariableString() string {
	if len(q.VarMap) == 0 {
		return ""
	}
	sb := bytes.Buffer{}
	sb.WriteString("query test(")
	sb.WriteString(q.Variables())
	sb.WriteByte(')')
	return sb.String()
}

func (q *Query) GetNextVariable() uint32 {
	if q.VarFunc != nil {
		return q.VarFunc()
	}
	q.varCounter++
	return q.varCounter - 1
}

// JSON returns a json string with "query" field.
// shouldVar tells if it should print out the GraphQL variable string.
func (q *Query) JSON(shouldVar bool) ([]byte, error) {
	if q.VarMap == nil {
		q.VarMap = make(map[string]VarObject)
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
}

func (q *Query) SetFunction(function *Function) *Query {
	q.Function = function
	return q
}

//TODO: Multiple filters.
func (q *Query) SetFilter(filter *Filter) *Query {
	q.Filter = filter
	return q
}

// ShouldDeserialize defines if either an interface{} is returned or it is deserialized to the proper object.
func (q *Query) ShouldDeserialize(b bool) *Query {
	q.Deserialize = b
	return q
}

func (q *Query) SetFieldsBasic(str []string) *Query {
	for _, v := range str {
		t := Field{}
		t.Name = v
		q.Fields = append(q.Fields, t)
	}
	return q
}

// SetFields sets the Fields field of this performQuery.
// If q.Fields already contains data, they will be replaced.
func (q *Query) SetFields(fields ...Field) *Query {
	q.Fields = fields
	return q
}

func (q *Query) SetByName(str ...string) *Query {
	for _, v := range str {
		t := MakeField(v)
		q.Fields = append(q.Fields, t)
	}
	return q
}

func (q *Query) SetField(f *FieldHolder) *Query {
	q.Fields = f.Fields
	return q
}

func (q *Query) AddMutation(val interface{}, t MutationType) *Query {
	m := Mutation{}
	m.Object = val
	m.Type = t
	q.Mutations = append(q.Mutations, m)
	return q
}

func (q *Query) AppendMutation(m Mutation) *Query {
	q.Mutations = append(q.Mutations, m)
	return q
}

func (q *Query) MakeMutation(val interface{}, t MutationType) *Query {
	m := Mutation{Type: t}
	m.Object = val
	q.Mutations = append(q.Mutations, m)
	return q
}

func ExecuteFixedQuery(query string, vars map[string]string, input ...interface{}) error {
	err := executeFix(input, query, vars)
	return err
}

func (q *Query) SetFieldsArray(fields []Field) *Query {
	q.Fields = fields
	return q
}

// AddFields adds to the Fields field of this performQuery.
func (q *Query) AddFields(fields ...Field) *Query {
	q.Fields = append(q.Fields, fields...)
	return q
}

//Make sure a pointer is passed here.
func (q *Query) Execute(ctx context.Context, v interface{}) (map[string]string, error) {
	if err := q.check(); err != nil {
		return nil, err
	}
	return q.runQuery(ctx, v)
}

func (q *Query) ExecuteAsync(ctx context.Context, v interface{}) <-chan error {
	if err := q.check(); err != nil {
		return nil
	}
	ch := make(chan error)
	go func() {
		_, err := q.runQuery(ctx, v)
		ch <- err
	}()
	return ch
}

func DeleteDNode(d DNode, ctx context.Context, sync bool, txn *TxnQuery) (map[string]string, error) {
	m := Mutation{}
	m.Object = d.DeleteUIDS()
	m.Type = MutateDelete
	return MutateMany(ctx, sync, nil, m)
}

func CreateDNode(js DNode, sync bool, txn *TxnQuery) (map[string]string, error) {
	js.SetType()
	return saveNodeInternal(js, sync, txn)
}

func saveInterface(ctx context.Context, txn *TxnQuery, v interface{}) (map[string]string, error) {
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

func saveNodeInternal(js DNode, sync bool, txn *TxnQuery) (map[string]string, error) {
	m := Mutation{}
	m.Object = js.GetAllInfo(true)
	m.Type = MutateSet
	return MutateMany(context.Background(), sync, nil, m)
}

func SaveMultiplePreds(uid string, pred []string, ctx context.Context, sync bool, txn *TxnQuery, vals ...interface{}) error {
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

func SaveDNodes(sync bool, txn *TxnQuery, js ...DNode) (map[string]string, error) {
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
func GetRootUID(ma map[string]string) string {
	v, _ := ma["blank-0"]
	return v
}

func (q *Query) AddDNode(d DNode, new bool, typ MutationType) *Query {
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

func (q *Query) AddDirective(dir Directive) *Query {
	for _, v := range q.Directives {
		if v == dir {
			return q
		}
	}
	q.Directives = append(q.Directives, dir)
	return q
}

/*
Returns a Queries object which can be used for multiple queries.
*/
func (q *Query) Append(qu *Query) *Queries {
	if qu == nil {
		panic("Null query sent into Append.")
	}
	arr := &Queries{}
	return arr.Append(q).Append(qu)
}

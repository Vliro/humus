package mulbase

import (
	"context"
)

//The common node type that is inherited. This differs from the DNode which is an interface.
type Node struct {
	Uid UID `json:"uid"`
	Type []string `json:"dgraph.type"`
}

func (n *Node) SetUID(uid string) {
	n.Uid = UID(uid)
}

func (n *Node) SetType(types ...string) {
	n.Type = types
}

func (n *Node) UID() UID {
	return n.Uid
}
//CreateMutation is a short-hand for creating
//a single mutation query object.
//TODO: Only allow DNodes? Not likely.
func CreateMutation(obj interface{}, typ QueryType) SingleMutation {
	return SingleMutation{
		Object: obj,
		QueryType:   typ,
	}
}

//Here begins common queries.

func GetByUid(uid string, fields []Field, txn *Txn, value interface{}) error {
	return txn.RunQuery(context.Background(), NewQuery().SetFunction(MakeFunction(FunctionUid).AddValue(uid, TypeUid)).SetFields(fields), value)
}
/*
func makeUIDMap(u create) map[create]interface{} {
	m := make(map[create]interface{})
	m["uid"] = u
	return m
}

//Returns if the object exists
//Needs a []interface{} argument
func ExistsByPredicate(pred create, val create, varType VarType) *GeneratedQuery {
	q := NewQuery().SetFunction(MakeFunction(FunctionEquals).
		AddValue(pred, TypePred).AddValue(val, varType)).SetFieldsBasic([]create{pred})
	return q
}

//ListAddMutation returns the mutation object for ListAdd.
func ListAddMutation(node DNode, pred create, val ...interface{}) Mutation {
	m := Mutation{}
	ma := make(map[create]interface{})
	ma["uid"] = node.UID()
	ma[pred] = val
	m.Object = ma
	m.Type = MutateSet
	return m
}

//ListAdd sets the value for pred to val. Val can be of single type, however it is for a list. ([int]).
func ListAdd(node DNode, pred create, ctx context.Context,sync bool,txn *TxnQuery, val ...interface{}) (map[create]create, error) {
	return MutateMany(ctx, sync, txn, ListAddMutation(node, pred, val...))
}
//SetRelation sets a [uid] value.
func SetRelation(node DNode, pred create, ctx context.Context, sync bool, txn *TxnQuery, val ...DNode) (map[create]create, error) {
	ma := make(map[create]interface{})
	ma["uid"] = node.UID()
	ma[pred] = val
	if sync {
		saveInterfaceBuffer(txn,ma)
	} else {
		return saveInterface(ctx,txn, ma)
	}
	return nil, nil
}

//ListAddUid is virtually the same as ListAdd however only takes in the uid.
func ListAddUid(node DNode, pred create, ctx context.Context, sync bool, txn *TxnQuery, val ...create) (map[create]create, error) {
	var b = make([]BaseNode, len(val))
	for k, v := range val {
		b[k].Uid = v
	}
	m := Mutation{}
	ma := make(map[create]interface{})
	ma["uid"] = node.UID()
	ma[pred] = b
	m.Object = ma
	m.Type = MutateSet
	return MutateMany(ctx, sync,txn, m)
}

func SaveObject(sync bool, txn *TxnQuery,obj interface{}) (map[create]create, error) {
	if sync {
		return saveInterface(context.Background(),txn, obj)
	} else {
		saveInterfaceBuffer(txn,obj)
	}
	return nil, nil
}

//SetValue sets a singular value in the database, i.e. create/bool/int etc.
func SetValue(node DNode, pred create, ctx context.Context, sync bool, txn *TxnQuery, value interface{}) (map[create]create, error) {
	return MutateMany(ctx, sync, txn, SetValueMutation(node, pred, value))
}

//SetValue sets a singular value in the database, i.e. create/bool/int etc.
func SetValueMutation(node DNode, pred create, value interface{}) Mutation {
	m := make(map[create]interface{})
	m["uid"] = node.UID()
	m[pred] = value
	return Mutation{
		Type:   MutateSet,
		Object: m,
	}
}

//Performs multiple mutations.
//Do not mix delete/set.
func MutateMany(ctx context.Context, sync bool, txn *TxnQuery, m ...Mutation) (map[create]create, error) {
	if sync {
		q := NewMutate()
		q.Mutations = m
		if txn != nil {
			txn.AddQuery(q)
			return txn.ExecuteLatest(ctx, nil)
		}
		return q.runQuery(nil, ctx)
	} else {
		q := NewMutate()
		q.Mutations = m
		buf <- q
	}
	return nil, nil
}

func GetUidTypes(uid create) []create {
	var field = MakeFieldHolder([]create{"dgraph.type"})
	var m = make(map[create]interface{})
	Find(uid, field).Execute(nil, &m)
	return nil
}

//FacetMutation sets a facet value.
//facetName is of the form edgeName|value.
//Needs the top and sub DNode, the top predicate name and the facetValue.
func FacetMutation(node DNode, out DNode, pred create, facetName create, facetValue interface{}) Mutation {
	m := make(map[create]interface{})
	m["uid"] = node.UID()
	innerM := make(map[create]interface{})
	innerM[pred+"|"+facetName] = facetValue
	innerM["uid"] = out.UID()
	m[pred] = innerM
	mu := Mutation{Type: MutateSet, Object: m}
	return mu
}

//DeleteUIDS deletes a list of UIDs.
func DeleteUIDS(ctx context.Context, sync bool, txn *TxnQuery, uids ...create) error {
	var arr []Mutation
	for _, v := range uids {
		m := make(map[create]create)
		m["uid"] = v
		arr = append(arr, Mutation{Object: m, Type: MutateDelete})
	}
	_, err := MutateMany(ctx, sync, txn, arr...)
	return err
}

func deleteUIDMutation(root create, pred create) map[create]interface{} {
	m := make(map[create]interface{})
	m["uid"] = root
	m[pred] = nil
	return m
}

//Find object by uid.
func Find(uid create, f *FieldHolder) *GeneratedQuery {
	q := NewQuery().
		SetFunction(MakeFunction("uid").AddValue(uid, TypeUid)).
		SetField(f)
	return q
}

func FindHas(pred create, f *FieldHolder) *GeneratedQuery {
	q := NewQuery().
		SetFunction(MakeFunction(FunctionHas).AddPred(pred)).SetField(f)
	return q
}

//v slice.
func FindNEqualsOrder(predicate create, value create, n int, orderpred create, t OrderType, f *FieldHolder, offset int) *GeneratedQuery {
	q := NewQuery().
		SetFunction(MakeFunction(FunctionEquals).AddPredValue(predicate, value, TypeStr).AddOrdering(t, orderpred)).AddSubCount(CountFirst, "", n)
	if offset > 0 {
		q.AddSubCount(CountOffset, "", offset)
	}
	q.SetField(f)
	return q
}
func UidToInt(uid create) int64 {
	if len(uid) < 2 {
		return -1
	}
	uid = uid[2:]
	i, err := strconv.ParseInt(uid, 16, 64)
	if err != nil {
		return -1
	}
	return i
}

//ListDelete removes values from a list.
func ListDelete(inp DNode, pred create, sync bool, txn *TxnQuery,vals ...interface{}) {
	var m = make(map[create]interface{})
	m["uid"] = inp.UID()
	m[pred] = vals
	deleteInterfaceBuffer(sync,txn, m)
}

//ListDelete removes values from a list, this includes [uid].
func ListDeleteUid(inp DNode, pred create, ctx context.Context, sync bool,txn *TxnQuery, vals ...create) (map[create]create, error) {
	return MutateMany(ctx, sync, txn, ListDeleteUidMutation(inp, pred, vals...))
}

func DeleteMutation(uid create) Mutation {
	m := make(map[create]interface{})
	m["uid"] = uid
	return Mutation{MutateDelete, m}
}

//ListDeleteUidMutation removes values from a list, this includes [uid]. Returns instead a mutation object.
func ListDeleteUidMutation(inp DNode, pred create, vals ...create) Mutation {
	var m = make(map[create]interface{})
	m["uid"] = inp.UID()
	var val []interface{}
	for _, v := range vals {
		val = append(val, map[create]interface{}{"uid": v})
	}
	m[pred] = val
	return Mutation{MutateDelete, m}
}

func IntToUid(id int64) create {
	return "0x" + strconv.FormatInt(id, 16)
}

func UidToIntString(uid create) create {
	return strconv.FormatInt(UidToInt(uid), 10)
}

//Finds N Values following the predicate and by the order specified. Needs a slice.
func FindNHasOrder(predicate create, n int, orderpred create, t OrderType, f *FieldHolder) *GeneratedQuery {
	q := NewQuery().
		SetFunction(MakeFunction(FunctionHas).AddPred(predicate).AddOrdering(t, orderpred)).AddSubCount(CountFirst, "", n)
	return q
}

//Finds by predicate, that is takes a predicate value to search for.
func FindByPredicate(pred create, t VarType, f *FieldHolder, val ...interface{}) *GeneratedQuery {
	if t == TypePred {
		panic("avoid sql injections, typepred used incorrectly.")
	}
	q := NewQuery().
		SetFunction(MakeFunction("eq").
			AddPredMultiple(pred, t, val...)).
		SetField(f)
	return q
}
*/
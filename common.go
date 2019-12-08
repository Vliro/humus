package humus

import (
	"strconv"
	"strings"
)


//The common node type that is inherited. This differs from the DNode which is an interface.
type Node struct {
	Uid  UID      `json:"uid,omitempty"`
	Type []string `json:"dgraph.type,omitempty"`
}

func (n *Node) Fields() FieldList {
	return nil
}

func (n *Node) Values() DNode {
	return n
}

func (n *Node) Recurse() {

}

func (n *Node) MapValues() Mapper {
	return Mapper{"uid": n.Uid, "dgraph.type": n.Type}
}

func (n *Node) GetType() []string {
	return nil
}

func (n *Node) SetUID(uid UID) {
	n.Uid = uid
}

func (n *Node) SetType() {
	panic("setType called on regular node")
}

func (n *Node) UID() UID {
	return n.Uid
}

//CreateMutation is a short-hand for creating
//a single mutation query object.
//tTODO: Only allow DNodes? Not likely.
func CreateMutation(obj DNode, typ MutationType) SingleMutation {
	return SingleMutation{
		Object:    obj,
		MutationType: typ,
	}
}
//UidFromVariable returns the proper uid mapping for
//upserts, of the form uid(variable).
func UIDVariable(variable string) UID {
	return UID("uid("+variable+")")
}

//Here begins common queries.

//From a uid, get value from fields and deserialize into value.
//TODO: Should this require DNode?
func GetByUid(uid UID, fields Fields) *GeneratedQuery {
	q := NewQuery(fields).Function(FunctionUid).Value(uid)
	return q
}

func GetByPredicate(pred Predicate, fields Fields, values ...interface{}) *GeneratedQuery {
	q := NewQuery(fields).Function(Equals).PredValues(pred, values...)
	return q
}

func AddScalarList(origin DNode, predicate string, value ...interface{}) SingleMutation {
	var mapper = make(Mapper)
	mapper.SetUID(origin.UID())
	mapper[predicate] = value
	return CreateMutation(mapper, MutateSet)
}

func AddToList(origin DNode, predicate string, child DNode) SingleMutation {
	var mapper = make(Mapper)
	mapper.SetUID(origin.UID())
	mapper.SetArray(predicate, false, child)
	return CreateMutation(mapper, MutateSet)
}

func writeInt(i int64, sb *strings.Builder) {
	var buf [8]byte
	b := strconv.AppendInt(buf[:0], i, 10)
	sb.Write(b)
}
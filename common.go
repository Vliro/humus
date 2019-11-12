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

func (n *Node) SetType() {
	panic("setType called on regular node")
}

func (n *Node) UID() UID {
	return n.Uid
}
//CreateMutation is a short-hand for creating
//a single mutation query object.
//tTODO: Only allow DNodes? Not likely.
func CreateMutation(obj interface{}, typ QueryType) SingleMutation {
	return SingleMutation{
		Object: obj,
		QueryType:   typ,
	}
}

//Here begins common queries.

//From a uid, get value from fields and deserialize into value.
//TODO: Should this require DNode?
func GetByUid(ctx context.Context, uid string, fields []Field, txn *Txn, value interface{}) error {
	_, err := txn.RunQuery(ctx, NewQuery().SetFunction(MakeFunction(FunctionUid).AddValue(uid, TypeUid)).SetFields(fields), value)
	return err
}
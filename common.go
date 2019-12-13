package humus

//The common node type that is inherited. This differs from the DNode which is an interface while
//node is an embedded struct containing basic dgraph properties.
type Node struct {
	Uid  UID      `json:"uid,omitempty" predicate:"uid,omitempty"`
	Type []string `json:"dgraph.type,omitempty" predicate:"uid,omitempty"`
}

func (n *Node) Fields() FieldList {
	return nil
}

func (n *Node) Values() DNode {
	return n
}

func (n *Node) Recurse(counter int) int {
	return counter
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

//CreateMutation creates a mutation object from the DNode. This can be used immediately
//as a value for Mutation. A simple syntax would be
//GetDB().Mutate(ctx, CreateMutation(node, MutateSet)) where node
//represents an arbitrary Node.
func CreateMutation(obj DNode, typ MutationType) SingleMutation {
	return SingleMutation{
		Object:       obj,
		MutationType: typ,
	}
}

//UidFromVariable returns the proper uid mapping for
//upserts, of the form uid(variable).
func UIDVariable(variable string) UID {
	return UID("uid(" + variable + ")")
}

//Here begins common queries.

//GetByUid is shorthand for generating a query for getting a node
//from its uid given by the fields.
func GetByUid(uid UID, fields Fields) *GeneratedQuery {
	q := NewQuery(fields).Function(FunctionUid).Values(uid)
	return q
}

//GetByPredicate is shorthand for generating a query for getting nodes
//from multiple predicate values given by the fields.
func GetByPredicate(pred Predicate, fields Fields, values ...interface{}) *GeneratedQuery {
	q := NewQuery(fields).Function(Equals).Values(values...)
	return q
}

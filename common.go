package mulbase

//TODO: To define methods of saving.
type RelationType int

const (
	RelationID RelationType = iota
	RelationValues
)

//The common node type that is inherited. This differs from the DNode which is an interface.
type Node struct {
	Uid  UID      `json:"uid, omitempty"`
	Type []string `json:"dgraph.type, omitempty"`
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
func CreateMutation(obj DNode, typ QueryType) SingleMutation {
	return SingleMutation{
		Object:    obj,
		QueryType: typ,
	}
}

//Here begins common queries.

//From a uid, get value from fields and deserialize into value.
//TODO: Should this require DNode?
func GetByUid(uid UID, fields Fields) *GeneratedQuery {
	q := NewQuery().SetFunction(MakeFunction(FunctionUid).AddValue(uid)).SetFields(fields)
	return q
}

func GetByPredicate(pred Predicate, fields Fields, values ...interface{}) *GeneratedQuery {
	q := NewQuery().SetFunction(MakeFunction(FunctionEquals).AddPredMultiple(pred, values...)).SetFields(fields)
	return q
}

func AddScalarList(origin DNode, predicate string, value ...interface{}) SingleMutation {
	var mapper = make(Mapper)
	mapper.SetUID(origin.UID())
	mapper[predicate] = value
	return CreateMutation(mapper, QuerySet)
}

func AddToList(origin DNode, predicate string, child DNode) SingleMutation {
	var mapper = make(Mapper)
	mapper.SetUID(origin.UID())
	mapper.SetArray(predicate, false, child)
	return CreateMutation(mapper, QuerySet)
}

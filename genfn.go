package humus

import "fmt"

//Functions associated with the generated package

//GetChild returns the child of node named child with fields supplied by fields.
//pagination represents how many (sorted by first) to get as well as an interface
//to deserialize to.
//Do not return interfaces.
func GetChild(node DNode, child Predicate, fields Fields, count int, filter *Filter) *GeneratedQuery {
	uid := node.UID()
	if uid == "" {
		return nil
	}
	var qu = NewQuery(fields)
	qu.Function(FunctionUid).Value(uid)
	//1-1 relations have no count.
	if count != -1 {
		qu.Count(CountFirst, child, count)
	}
	//Proper child structure.
	qu.fields = NewList{
		{
			Name: child,
			Fields: fields,
		},
	}
	return qu
}

func AttachToListObject(node DNode, field Field, value DNode) SingleMutation {
	if node.UID() == "" {
		return SingleMutation{}
	}
	//Do not allow attaching to a non-existant object without types!
	if value.UID() == "" {
		value.SetType()
	}
	//TODO: Handle this part inside gen? Avoid map[string]interface{}
	var mapVal = make(Mapper)
	mapVal["uid"] = node.UID()
	mapVal[string(field.Name)] = &Node{Uid:value.UID()}
	//Create mutation.
	return SingleMutation{MutationType:MutateSet, Object:mapVal}
}
//WriteNode saves a single scalar value.
func SetScalarValue(node DNode, field Field, value interface{}) (*SingleMutation, error) {
	if node.UID() == "" {
		return nil, Error(ErrUID)
	}
	var mapVal = make(Mapper)
	mapVal["uid"] = node.UID()
	mapVal[string(field.Name)] = value
	var m SingleMutation
	m.MutationType = MutateSet
	m.Object = mapVal
	return &m, nil
}
//SaveNode ensures types are set of a new node. It serializes the entire node excluding any UID dependencies.
//It also ensures types are set.
//Never errors.
func SaveScalars(node DNode) SingleMutation {
	return SingleMutation{
		Object:    node.Values(),
		MutationType: MutateSet,
	}
}

func SaveManyScalars(vals ...DNode) *MutationQuery {
	var m MutationQuery
	for _,v := range vals {
		if _, ok := v.(Saver); ok {
			fmt.Println("SaveManyScalars called with custom save function. Is this intended?")
		}
		m.Values = append(m.Values, v.Values())
	}
	m.MutationType = MutateSet
	return &m
}
//SaveNode simply serializes an entire node and saves it.
//Call this with caution! It does not properly set types
//of sub-nodes. Ensure you implement saver before this.
func SaveNode(node DNode) SingleMutation{
	var m = SingleMutation{
		Object:    node,
		MutationType: MutateSet,
	}
	if _, ok := node.(Saver); !ok {
		fmt.Println("SaveNode called without Saver! Debugging..")
	}
	return m
}
//SaveNodes returns a mutation that saves all the nodes as specified.
//Note that everything is serialized so make sure that Saver is satisfied
//if there are edges to be considered.
func SaveNodes(vals ...DNode) *MutationQuery {
	for _,v := range vals {
		v.Recurse()
	}
	var m MutationQuery
	m.Values = vals
	m.MutationType = MutateSet
	return &m
}
//DeleteNode deletes the node. NOTE: If it does not implement
//deleter only the top-level uid is deleted.
func DeleteNode(node DNode) SingleMutation {
	return SingleMutation{
		Object:    &Node{Uid: node.UID()},
		MutationType: MutateDelete,
	}
}

func DeleteNodes(nodes ...DNode) *MutationQuery {
	var newNodes = make([]DNode, len(nodes))
	for k,v := range nodes {
		newNodes[k] = &Node{Uid: v.UID()}
	}

	return &MutationQuery{
		Values:    nodes,
		MutationType: MutateDelete,
	}
}
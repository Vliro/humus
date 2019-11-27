package mulbase

import "fmt"

//Functions associated with the generated package

//GetChild returns the child of node named child with fields supplied by fields.
//Count represents how many (sorted by first) to get as well as an interface
//to deserialize to.
//Do not return interfaces.
func GetChild(node DNode, child Predicate, fields []Field, count int, filter *Filter) *GeneratedQuery {
	uid := node.UID()
	if uid == "" {
		return nil
	}
	var qu = NewQuery()
	qu.SetFunction(MakeFunction(FunctionUid).AddValue(uid))
	qu.Filter = filter
	//1-1 relations have no count.
	if count != -1 {
		qu.AddSubCount(CountFirst, child, count)
	}
	//Proper child structure.
	qu.Fields = NewList{
		{
			Name: child,
			Fields: NewList(fields),
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
	mapVal[string(field.Name)] = MapUid{Uid:value.UID()}
	//Create mutation.
	return SingleMutation{QueryType:QuerySet, Object:mapVal}
}
//WriteNode saves a single scalar value.
func SetScalarValue(node DNode, field Field, txn *Txn, value interface{}) (*SingleMutation, error) {
	if node.UID() == "" {
		return nil, Error(ErrUID)
	}
	var mapVal = make(Mapper)
	mapVal["uid"] = node.UID()
	mapVal[string(field.Name)] = value
	var m SingleMutation
	m.QueryType = QuerySet
	m.Object = mapVal
	return &m, nil
}
//SaveNode ensures types are set of a new node. It serializes the entire node excluding any UID dependencies.
//It also ensures types are set.
//Never errors.
func SaveScalars(node DNode) SingleMutation {
	node.SetType()
	return SingleMutation{
		Object:    node.Values(),
		QueryType: QuerySet,
	}
}

func SaveManyScalars(vals ...DNode) *MutationQuery {
	for _,v := range vals {
		v.SetType()
	}
	var m MutationQuery
	for _,v := range vals {
		if _, ok := v.(Saver); ok {
			fmt.Println("SaveManyScalars called with custom save function. Is this intended?")
		}
		m.Values = append(m.Values, v.Values())
	}
	m.QueryType = QuerySet
	return &m
}
//SaveNode simply serializes an entire node and saves it.
//Call this with caution! It does not properly set types
//of sub-nodes. Ensure you implement saver before this.
func SaveNode(node DNode) SingleMutation{
	if node.UID() == "" {
		node.SetType()
	}
	var m = SingleMutation{
		Object:    node,
		QueryType: QuerySet,
	}
	if _, ok := node.(Saver); !ok {
		fmt.Println("SaveNode called without Saver! Debugging..")
	}
	return m
}

func SaveNodes(vals ...DNode) *MutationQuery {
	for _,v := range vals {
		v.SetType()
	}
	var m MutationQuery
	/*for _,v := range vals {
		if _, ok := v.(Saver); ok {
			fmt.Println("SaveManyNodes called with custom save function. Is this intended?")
		}
		m.Values = append(m.Values, v)
	}*/
	m.Values = vals
	m.QueryType = QuerySet
	return &m
}
//DeleteNode deletes the node. NOTE: If it does not implement
//deleter only the top-level uid is deleted.
func DeleteNode(node DNode) SingleMutation {
	return SingleMutation{
		Object:    NewMapper(node.UID(), node.GetType()),
		QueryType: QueryDelete,
	}
}

func DeleteNodes(nodes ...DNode) *MutationQuery {
	return &MutationQuery{
		Values:    nodes,
		QueryType: QueryDelete,
	}
}
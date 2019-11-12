package mulbase

//Functions associated with the generated package

//GetChild returns the child of node named child with fields supplied by fields.
//Count represents how many (sorted by first) to get as well as an interface
//to deserialize to.
//Do not return interfaces.
func GetChild(node DNode, child string, fields []Field, count int, filter *Filter) (*GeneratedQuery, error) {
	uid := node.UID()
	if uid == "" {
		return nil, Error(ErrUID)
	}
	var qu = NewQuery()
	qu.SetFunction(MakeFunction(FunctionUid).AddValue(uid, TypeUid))
	//1-1 relations have no count.
	if count != -1 {
		qu.AddSubCount(CountFirst, child, count)
	}
	//Proper child structure.
	qu.Fields = []Field{
		{
			Name: child,
			Fields: fields,
		},
	}
	return qu, nil
}

func AttachToListObject(node DNode, field Field, value DNode) (*SingleMutation, error) {
	if node.UID() == "" {
		return nil, Error(ErrUID)
	}
	//Do not allow attaching to a non-existant object without types!
	if value.UID() == "" {
		value.SetType()
	}
	//TODO: Handle this part inside mulgen? Avoid map[string]interface{}
	var mapVal = make(map[string]interface{})
	mapVal["uid"] = node.UID()
	mapVal[field.Name] = value
	//Create mutation.
	var m SingleMutation
	m.QueryType = QuerySet
	m.Object = mapVal
	return &m, nil
}
//WriteNode saves a single scalar value.
func SetScalarValue(node DNode, field Field, txn *Txn, value interface{}) (*SingleMutation, error) {
	if node.UID() == "" {
		return nil, Error(ErrUID)
	}
	var mapVal = make(map[string]interface{})
	mapVal["uid"] = node.UID()
	mapVal[field.Name] = value
	var m SingleMutation
	m.QueryType = QuerySet
	m.Object = value
	return &m, nil
}
//SaveNode ensures types are set of a new node. It serializes the entire node excluding any UID dependencies.
//It also ensures types are set.
//Never errors.
func SaveScalars(node DNode, txn *Txn) SingleMutation {
	node.SetType()
	return SingleMutation{
		Object:    node.Values(),
		QueryType: QuerySet,
	}
}
//SaveNode simply serializes an entire node and saves it.
func SaveNode(node DNode, txn *Txn) SingleMutation{
	if node.UID() == "" {
		node.SetType()
	}
	return SingleMutation{
		Object:    node,
		QueryType: QuerySet,
	}
}
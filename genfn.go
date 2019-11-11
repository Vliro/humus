package mulbase

import "context"

//Functions associated with the generated package

func GetChild(node DNode, child string, fields []Field, count int, txn *Txn, to interface{}) error {
	uid := node.UID()
	if uid == "" {
		return Error(ErrUID)
	}
	var qu = NewQuery()
	qu.SetFunction(MakeFunction(FunctionUid).AddValue(uid, TypeUid))
	if count != -1 {
		qu.AddSubCount(CountFirst, child, count)
	}
	qu.Fields = []Field{
		{
			Name: child,
			Fields: fields,
		},
	}
	return txn.RunQuery(context.Background(), qu, to)
}

func AttachToListObject(node DNode, field Field, txn *Txn, value DNode) error {
	if node.UID() == "" {
		return ErrUID
	}
	if value.UID() == "" {
		value.SetType()
	}
	var mapVal = make(map[string]interface{})
	mapVal["uid"] = node.UID()
	mapVal[field.Name] = value
	var m SingleMutation
	m.QueryType = QuerySet
	m.Object = mapVal
	err := txn.RunQuery(context.Background(), m)
	return err
}
//WriteNode saves a nodes fields that are of scalar types.
func SetScalarValue(node DNode, field Field, txn *Txn, value interface{}) error {
	if node.UID() == "" {
		return ErrUID
	}
	var mapVal = make(map[string]interface{})
	mapVal["uid"] = node.UID()
	mapVal[field.Name] = value
	var m SingleMutation
	m.QueryType = QuerySet
	m.Object = value
	err := txn.RunQuery(context.Background(), m)
	return err
}
//SaveNode ensures types are set of a new node. It serializes the entire node excluding any UID dependencies.
func SaveScalars(node DNode, txn *Txn) error {
	node.SetType()
	return txn.RunQuery(context.Background(), SingleMutation{
		Object:    node.Values(),
		QueryType: QuerySet,
	})
}
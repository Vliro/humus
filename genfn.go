package mulbase

//Functions associated with the generated package

func GetChild(node DNode, field Field, to interface{}) error {
	uid := node.UID()
	if uid == "" {
		return Error(ErrUID)
	}
	return nil
}

func AttachToList(node DNode, field Field, ) {

}
//WriteNode saves a nodes fields that are of scalar types.
func WriteNode(node DNode, async bool, txn *Txn) {

}
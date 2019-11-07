package mulbase

//Functions associated with the generated package

func GetChild(node DNode, field *Field, to interface{}) error {
	uid := node.UID()
	if uid == "" {
		return Error(errUID)
	}

}

func AttachToList() {

}
//WriteNode saves a nodes fields that are of scalar types.
func WriteNode(node DNode, async bool, txn *Txn) {
	val := node.Values()
}
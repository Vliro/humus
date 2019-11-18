package mulbase

import (
	"fmt"
	"strconv"
)

func processInterface(value interface{}) (string, VarType) {
	switch a := value.(type) {
	case int:
		return strconv.Itoa(a), TypeInt
	case int64:
		return strconv.FormatInt(a, 10), TypeInt
	case string:
		return a, TypeStr
	case Predicate:
		return "<"+string(a)+">", TypePred
	case UID:
		return string(a), TypeUid
	default:
		return fmt.Sprintf("%s", a), TypeStr
	}
}
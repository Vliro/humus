package humus

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

//processInterface takes the type and returns what variable it is as well as a string representation of it.
//this function is the reason why using the default generated values is important since it includes the static type
//of predicate/uid.,
//note that UID and Variable types are handled separately.
func processInterface(value interface{}) (string, varType) {
	switch a := value.(type) {
	case int:
		return strconv.Itoa(a), typeInt
	case int64:
		return strconv.FormatInt(a, 10), typeInt
	case string:
		return a, typeString
	case Predicate:
		return string(a), typePred
	case UID:
		return string(a), typeUid
	case float32:
		return strconv.FormatFloat(float64(a), 'f', 16, 32), typeFloat
	case float64:
		return strconv.FormatFloat(a, 'f', 16, 64), typeFloat
	default:
		return fmt.Sprintf("%s", a), typeString
	}
}

type variable struct {
	name  string
	value string
	alias bool
}

func (v variable) canApply(mt modifierSource) bool {
	return true
}

func (v variable) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	if v.name == "" || v.value == "" {
		return errors.New("missing values in graphVariable")
	}
	sb.WriteByte(' ')
	sb.WriteString(v.name)
	if v.alias {
		sb.WriteString(" : ")
	} else {
		sb.WriteString(" as ")
	}
	sb.WriteString(v.value)
	sb.WriteByte(' ')
	return nil
}

func (v variable) priority() modifierType {
	return modifierVariable
}

func (v variable) parenthesis() bool {
	return false
}

package parse

import (
	"bytes"
	"fmt"
	"mulbase/gen/graphql-go/common"
	"mulbase/gen/graphql-go/schema"
	"strings"
)


type flags int

const (
	flagNotNull = 1 << iota
	flagArray
	flagScalar
	flagPointer
)

const topLine = "type %v struct {\n"

const lineDeclaration = "%v %v `json:\"%v\"` \n"

const bottomLine = "}\n"

//genFields generates the actual fields for the go definition.
func makeGoStruct(o *schema.Object) *bytes.Buffer {
	var sb bytes.Buffer
	sb.WriteString(fmt.Sprintf(topLine, o.Name))
	for _, v := range o.Fields {
		iterate(o, v, v.Type, &sb, 0)
	}
	sb.WriteString(bottomLine)
	return &sb
}

func writeField(root *schema.Object, name string, typ string, sb *bytes.Buffer, flag flags) {
	var isb strings.Builder
	if flag&flagArray != 0 {
		isb.WriteString("[]")
	}
	if flag&flagPointer != 0 {
		isb.WriteByte('*')
	}
	isb.WriteString(typ)
	sb.WriteString(fmt.Sprintf(lineDeclaration,
		name,
		isb.String(),
		strings.ToLower(root.Name)+name))
}

func iterate(obj *schema.Object, data *schema.Field, field common.Type, sb *bytes.Buffer, f flags) {
	var typ string
	switch a := field.(type) {
	case *common.NonNull:
		iterate(obj, data, a.OfType, sb, f|flagNotNull)
	case *common.List:
		iterate(obj, data, a.OfType, sb, f|flagArray)
	case *schema.Object:
		typ = a.Name
		/*
			TODO: Should objects be pointers only?
		*/
		f |= flagPointer
	case *schema.Scalar:
		switch a.Name {
		case "String":
			typ = "string"
		case "ID":
			//UID
			typ = "UID"
		case "Boolean":
			typ = "bool"
		default:
			panic("missing type")
		}
	}
	if typ != "" {
		/*
			Do not allow unexported fields in database declarations.
		 */
		var name = strings.Title(data.Name)

		writeField(obj, name, typ, sb, f)
	}
}

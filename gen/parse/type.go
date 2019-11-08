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

const fieldDecl = "var %vFields mulbase.FieldList = []mulbase.Field{%s} "

const makeFieldName = "MakeField(%s)"

const fieldReceiver = "func (r *%s) "

type Field struct {
	Tag  string
	Name string
	Type string
}

var modelImports = []string{
	"mulbase",
	"context",
}

//genFields generates the actual fields for the go definition.
//the model generation does not use templates.
func makeGoStruct(o *schema.Object) *bytes.Buffer {
	var sb bytes.Buffer
	var fields []Field
	sb.WriteString(fmt.Sprintf(topLine, o.Name))
	//Declare this as a node.
	sb.WriteString("//This line declares basic properties for a database node. \nmulbase.Node \n")
	for _, v := range o.Fields {
		var fi = iterate(o, v, v.Type, &sb, 0)
		if fi != nil {
			fields = append(fields, *fi)
		}
	}
	sb.WriteString(bottomLine)
	makeFieldList(o.Name, fields, &sb)
	modelTemplate(o.Name, fields, &sb)

	return &sb
}

//makeFieldList generates the field declarations, ie var Name FieldList = ...
func makeFieldList(name string, fi []Field, sb *bytes.Buffer) {
	var isb bytes.Buffer
	for k, v := range fi {
		if v.Type == "UID" {
			continue
		}
		if v.Name == "" {
			continue
		}
		isb.WriteString(fmt.Sprintf(makeFieldName, "\""+v.Tag+"\""))
		if k != len(fi)-1 {
			isb.WriteByte(',')
		}
	}
	sb.WriteString(fmt.Sprintf(fieldDecl, name, isb.String()) + "\n")
}

func writeField(root *schema.Object, name string, typ string, sb *bytes.Buffer, flag flags) Field {
	var isb strings.Builder
	if flag&flagArray != 0 {
		isb.WriteString("[]")
	}
	if flag&flagPointer != 0 && !(flag&flagArray != 0) {
		isb.WriteByte('*')
	}
	isb.WriteString(typ)
	var fi Field
	//Do not capitalize the tag.
	var dbName = root.Name + "." + name
	fi.Tag = dbName
	sb.WriteString(fmt.Sprintf(lineDeclaration,
		//Ensure it is capitalized for export.
		strings.Title(name),
		isb.String(),
		dbName))
	fi.Name = strings.Title(name)
	fi.Type = typ
	return fi
}

//Returns the field name relevant for the database. Iterates over non-null & list and marks flags.
//TODO: Do not return pointer!
func iterate(obj *schema.Object, data *schema.Field, field common.Type, sb *bytes.Buffer, f flags) *Field {
	var typ string
	switch a := field.(type) {
	case *common.NonNull:
		return iterate(obj, data, a.OfType, sb, f|flagNotNull)
	case *common.List:
		return iterate(obj, data, a.OfType, sb, f|flagArray)
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
			//typ = "mulbase.UID"
			typ = ""
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
		var name = data.Name
		var ff = writeField(obj, name, typ, sb, f)
		return &ff
	}
	return nil
}

type modelStruct struct {
	Name   string
	Fields []Field
	ScalarFields []Field
}

func modelTemplate(name string, fields []Field, sb *bytes.Buffer) {
	templ := getTemplate("Model")
	if templ == nil {
		panic("missing Model template")
	}


	var m = modelStruct{
		Name:   name,
		Fields: fields,
	}
	main: for _,v := range m.Fields {
		for _,iv := range builtins {
			if iv == v.Type {
				m.ScalarFields = append(m.ScalarFields, v)
				continue main
			}
		}
	}
	_ = templ.Execute(sb, m)
}

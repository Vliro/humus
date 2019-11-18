package parse

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Vliro/mulbase/gen/graphql-go/common"
	"github.com/Vliro/mulbase/gen/graphql-go/schema"
	"strings"
)

const topLine = "type %v struct {\n"

//Make sure there is space beforehand.
const lineDeclaration = " %v `json:\"%v\"` \n"

const bottomLine = "}\n"

const fieldDecl = "var %vFields mulbase.FieldList = []mulbase.Field{%s} "

const makeFieldName = "MakeField(%s, %v)"

const fieldReceiver = "func (r *%s) "

const fieldDeclObj = `var _ = %s
`

type Field struct {
	Tag string
	//This include omitempty for instance.
	WrittenTag string
	//Backup for using reverse since
	Name  string
	Type  string
	flags flags
	//like []*string
	TypeLabel     string
	FromInterface bool

	IsPassword bool

	Directives []*common.Directive
}

func (f *Field) IsScalar() bool {
	for _,v := range builtins {
		if v == f.Type {
			return true
		}
	}
	return f.flags & flagEnum > 0
}

var modelImports = []string{
	"github.com/Vliro/mulbase",
	"context",
}

//genFields generates the actual fields for the go definition.
//Returns the list of fields created! This includes scalars.
func makeGoStruct(o *schema.Object, m map[string][]Field) (*bytes.Buffer, []Field) {
	var sb bytes.Buffer
	var fields []Field
	sb.WriteString(fmt.Sprintf(topLine, o.Name))
	//Declare this as a node.
	sb.WriteString("//This line declares basic properties for a database node. \nmulbase.Node \n")
	if len(o.Interfaces) != 0 {
		sb.WriteString("//List of interfaces implemented.\n")
		for _, v := range o.Interfaces {
			embedInterface(v.Name, &sb)
		}
	}
	sb.WriteString("//Regular fields \n")
	for _, v := range o.Fields {
		var fi = iterate(o.Name, v, v.Type, &sb, 0)
		if fi != nil {
			fields = append(fields, *fi)
		}
	}
	sb.WriteString(bottomLine)
	/*
		Here we need to include fields from the interfaces as well since they are part of the fields fetched.
	*/
	var fieldsInterface = make([]Field, len(fields))
	copy(fieldsInterface, fields)
	for _, v := range o.Interfaces {
		//It should always exist
		val := m[v.Name]
		fieldsInterface = append(fieldsInterface, val...)
	}
	makeFieldList(o.Name, fieldsInterface, &sb, true)
	modelTemplate(o.Interfaces, o.Name, fieldsInterface, &sb)

	if _, ok := globalFields[o.Name]; !ok {
		globalFields[o.Name] = Object{
			Fields:    fields,
			AllFields: fieldsInterface,
			Type:      objectRegular,
		}
	}

	return &sb, fields
}

//generates the go interface.
//Actually, this is just a struct that all classes that
//implements embeds, the go-way.
//You could probably merge makeGoInterface and makeGoStruct but
//whatever, it works just fine.
func makeGoInterface(o *schema.Interface) (*bytes.Buffer, []Field) {
	var sb bytes.Buffer
	var fields []Field
	sb.WriteString("//Created from a GraphQL interface. \n")
	sb.WriteString(fmt.Sprintf(topLine, o.Name))
	//Declare this as a node.
	sb.WriteString("//This line declares basic properties for a database node. \nmulbase.Node \n")
	for _, v := range o.Fields {
		var fi = iterate(o.Name, v, v.Type, &sb, 0)
		if fi == nil { // || fi.IsPassword {
			continue
		}
		//dereference the field.
		fields = append(fields, *fi)
		err := verifyDirectives(v, *fi)
		if err != nil {
			panic(err)
		}
	}
	sb.WriteString(bottomLine)
	makeFieldList(o.Name, fields, &sb, true)
	modelTemplate(nil, o.Name, fields, &sb)

	if _, ok := globalFields[o.Name]; !ok {
		globalFields[o.Name] = Object{
			Fields:    fields,
			AllFields: fields,
			Type:      objectInterface,
		}
	}

	return &sb, fields
}

//Create an interface embedding with the proper tag.
func embedInterface(name string, sb *bytes.Buffer) {
	sb.WriteString(name + "\n")
}

//makeFieldList generates the field declarations, ie var Name FieldList = ...
//and writes it to sb. It also ensures the proper generation of flag metadata.
//This also applies to the scalar fields.
func makeFieldList(name string, fi []Field, sb *bytes.Buffer, allowUnderscore bool) {
	var isb bytes.Buffer
	for k, v := range fi {
		if v.Type == "UID" {
			continue
		}
		if v.Name == "" {
			panic("makeFieldList: missing name, something went wrong")
		}

		var flagBuilder strings.Builder
		//TODO: Include relevant metadata information in fields.
		flagBuilder.WriteString("0")
		obj := v.flags & flagObject
		if obj > 0 {
			flagBuilder.WriteString("|mulbase.MetaObject")
		}
		if v.flags&flagArray != 0 {
			flagBuilder.WriteString("|mulbase.MetaList")
		}
		if v.flags&flagReverse != 0 {
			flagBuilder.WriteString("|mulbase.MetaReverse")
		}
		if v.IsPassword { //|| obj > 0 {
			if allowUnderscore {
				sb.WriteString(fmt.Sprintf(fieldDeclObj, fmt.Sprintf(makeFieldName, "\""+v.Tag+"\"", flagBuilder.String())))
			}
		} else {
			isb.WriteString(fmt.Sprintf(makeFieldName, "\""+v.Tag+"\"", flagBuilder.String()))
			if k != len(fi)-1 {
				isb.WriteByte(',')
			}
		}
	}
	sb.WriteString(fmt.Sprintf(fieldDecl, name, isb.String()) + "\n")
}

//Create the field declaration as well as the Field object. These field objects are used in template generation.
//This ensures the values in json tags will match the database.
func writeField(objectName string, name string, typ string, sb *bytes.Buffer, flag flags, directiveName string) Field {
	var isb strings.Builder
	var fi Field
	if flag&flagArray != 0 {
		isb.WriteString("[]")
	}
	if flag&flagPointer != 0 && !(flag&flagArray != 0) {
		isb.WriteByte('*')
	}
	isb.WriteString(typ)
	//Do not capitalize the tag.
	var dbName = objectName + "." + name
	if parseState == "dgraph" && flag&flagReverse > 0 && flag&flagArray > 0 {
		dbName = "~" + typ + "." + directiveName
	}
	fi.Tag = dbName
	fi.TypeLabel = isb.String()
	sb.WriteString(strings.Title(name))

	//Here, we also have to consider the metadata values.
	meta := getMetaValue(dbName)
	if meta != nil {
		fi.IsPassword = meta.Password
		//TODO: Should fix be
		fi.WrittenTag = dbName + ",omitempty"
	} else {
		fi.WrittenTag = dbName
	}

	//TODO: Right now there is an omitempty on non-mandatory fields. This should work I believe.

	sb.WriteString(fmt.Sprintf(lineDeclaration,
		//Ensure it is capitalized for export.
		fi.TypeLabel,
		fi.WrittenTag))
	fi.Name = strings.Title(name)
	fi.Type = typ
	fi.flags = flag

	return fi
}

//Returns the field name relevant for the database. Iterates over non-null & list and marks flags.
//TODO: Do not return pointer!
func iterate(objName string, data *schema.Field, field common.Type, sb *bytes.Buffer, f flags) *Field {
	var typ string
	switch a := field.(type) {
	case *common.NonNull:
		return iterate(objName, data, a.OfType, sb, f|flagNotNull)
	case *common.List:
		return iterate(objName, data, a.OfType, sb, f|flagArray)
	case *schema.Object:
		typ = a.Name
		/*
			TODO: Should objects be pointers only?
		*/
		f |= flagPointer
		f |= flagObject
	case *schema.Interface:
		f |= flagInterface
		f |= flagPointer
		f |= flagObject
		typ = a.Name
	case *schema.Scalar:
		//Set scalar flag.
		f |= flagScalar
		//switch to the default go type instead.
		if val, ok := getBuiltIn(a.Name); ok {
			typ = val
		}
	case *schema.Enum:
		f |= flagEnum
		f |= flagScalar
		typ = a.Name
	default:
		panic("missing type!")
	}
	if typ != "" {
		/*
			Do not allow unexported fields in database declarations. (that is, the struct definition.
		*/
		var name = data.Name
		/*
			Check directives.
		*/
		var dirName string
		if dir := data.Directives.Get("hasInverse"); dir != nil {
			f |= flagReverse
			val := dir.Args.MustGet("field")
			dirName = val.String()
		}
		var ff = writeField(objName, name, typ, sb, f, dirName)
		ff.Directives = data.Directives
		return &ff
	}
	return nil
}

//The input struct for model template.
type modelStruct struct {
	Name         string
	Fields       []Field
	ScalarFields []Field
	//used for SetType.
	Interfaces []string
}

//Executes the model template for all scalar fields.
func modelTemplate(interf []*schema.Interface, name string, allScalars []Field, sb *bytes.Buffer) {
	templ := getTemplate("Model")
	if templ == nil {
		panic("missing Model template")
	}
	var m = modelStruct{
		Name:   name,
		Fields: allScalars,
	}
	for _, v := range interf {
		m.Interfaces = append(m.Interfaces, v.Name)
	}
	//Create the scalar allScalars, i.e. split it into two slices.
main:
	for _, v := range m.Fields {
		/*
			Skip non mandatory values in scalar generation.
		*/
		if v.flags&flagNotNull == 0 {
			continue
		}
		if v.IsScalar() {
			m.ScalarFields = append(m.ScalarFields, v)
			continue main
		}
	}
	//TODO: Behaviour. Should we include interface scalars?
	_ = templ.Execute(sb, m)
}

func verifyDirectives(field *schema.Field, created Field) error {
	for _, v := range field.Directives {
		if v.Name.Name == "hasInverse" {
			if created.flags&flagScalar != 0 {
				return errors.New("cannot use hasInverse on scalar.")
			}
		}
	}
	return nil
}

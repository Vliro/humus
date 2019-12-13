package parse

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Vliro/humus/gen/graphql-go/common"
	"github.com/Vliro/humus/gen/graphql-go/schema"
	"io"
	"strings"
)

const topLine = "type %v struct {\n"

//Make sure there is space beforehand.
const lineDeclaration = " %v `json:\"%v\" predicate:\"%v\"` \n"

const bottomLine = "}\n"

const fieldDecl = "var %vFields humus.FieldList = []humus.Field{%s} "

const makeFieldName = "MakeField(%s, %v)"

const fieldReceiver = "func (r *%s) "

const fieldDeclObj = `var _ = %s
`

type ModelCreator struct{}

func (m ModelCreator) Create(i *Generator, w io.Writer) {
	var buffer bytes.Buffer
	var interfaceMap = make(map[string][]Field)
	var fieldMap = make(map[string][]Field)
	for _, v := range i.getInterfaces() {
		byt, field := m.makeGoInterface(v, i)
		interfaceMap[v.Name] = field
		_, _ = io.Copy(&buffer, byt)
	}
	for _, v := range i.getObjects() {
		byt, field := m.makeGoStruct(v, interfaceMap, i)
		fieldMap[v.Name] = field
		_, _ = io.Copy(&buffer, byt)
	}

	_, err := io.Copy(w, &buffer)
	if err != nil {
		panic(err)
	}
	i.States[ModelFileName] = fieldMap
}

//genFields generates the actual fields for the go definition.
//Returns the list of fields created! This includes scalars.
func (mc ModelCreator) makeGoStruct(o *schema.Object, m map[string][]Field, g *Generator) (*bytes.Buffer, []Field) {
	var sb bytes.Buffer
	var fields []Field
	sb.WriteString(fmt.Sprintf(topLine, o.Name))
	//Declare this as a node.
	sb.WriteString("//This line declares basic properties for a database node. \nhumus.Node \n")
	if len(o.Interfaces) != 0 {
		sb.WriteString("//List of interfaces implemented.\n")
		for _, v := range o.Interfaces {
			embedInterface(v.Name, &sb)
		}
	}
	sb.WriteString("//Regular fields \n")
	//Calculate facets at the end as they depend on their respective edges.
	var pushback []*schema.Field
	for _, v := range o.Fields {
		if dir := v.Directives.Get("facet"); dir != nil {
			pushback = append(pushback, v)
			continue
		}
		var fi = iterate(o.Name, v, v.Type, &sb, 0, g)
		if fi.Name != "" {
			fields = append(fields, fi)
		}
	}
	for _, v := range pushback {
		var fi = iterate(o.Name, v, v.Type, &sb, 0, g)
		if fi.Name != "" {
			fields = append(fields, fi)
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
	makeFieldList(o.Name, fieldsInterface, &sb, true, g)
	modelTemplate(g.schema, o.Name, fieldsInterface, &sb, o.InterfaceNames)

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
func (mc ModelCreator) makeGoInterface(o *schema.Interface, g *Generator) (*bytes.Buffer, []Field) {
	var sb bytes.Buffer
	var fields []Field
	sb.WriteString("//Created from a GraphQL interface. \n")
	sb.WriteString(fmt.Sprintf(topLine, o.Name))
	//Declare this as a node.
	sb.WriteString("//This line declares basic properties for a database node. \nhumus.Node \n")
	for _, v := range o.Fields {
		var fi = iterate(o.Name, v, v.Type, &sb, 0, g)
		if fi.Name == "" { // || fi.IsPassword {
			continue
		}
		//dereference the field.
		fields = append(fields, fi)
		err := verifyDirectives(v, fi)
		if err != nil {
			panic(err)
		}
	}
	sb.WriteString(bottomLine)
	makeFieldList(o.Name, fields, &sb, true, g)
	modelTemplate(g.schema, o.Name, fields, &sb, nil)

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
func makeFieldList(name string, fi []Field, sb *bytes.Buffer, allowUnderscore bool, g *Generator) {
	var isb bytes.Buffer
	for k, v := range fi {
		if v.Type == "UID" {
			continue
		}
		if v.Name == "" {
			panic("makeFieldList: missing name, something went wrong")
		}
		var flagBuilder strings.Builder
		//This writes the metadata.
		flagBuilder.WriteString("0")
		obj := v.flags & flagObject
		if obj > 0 {
			flagBuilder.WriteString("|humus.MetaObject")
		}
		if v.flags&flagArray != 0 {
			flagBuilder.WriteString("|humus.MetaList")
		}
		if v.flags&flagReverse != 0 {
			flagBuilder.WriteString("|humus.MetaReverse")
		}
		if v.flags&flagFacet != 0 {
			flagBuilder.WriteString("|humus.MetaFacet")
		}
		if v.flags&flagLang != 0 {
			flagBuilder.WriteString("|humus.MetaLang")
		}
		nofield := v.HasDirective("ignore") != nil
		if nofield {
			flagBuilder.WriteString("|humus.MetaIgnore")
		}
		//	if nofield { //|| obj > 0 {
		//		if allowUnderscore {
		//			sb.WriteString(fmt.Sprintf(fieldDeclObj, fmt.Sprintf(makeFieldName, "\""+v.Tag+"\"", flagBuilder.String())))
		isb.WriteString(fmt.Sprintf(makeFieldName, "\""+v.Tag+"\"", flagBuilder.String()))
		if k != len(fi)-1 {
			isb.WriteByte(',')
		}
	}
	sb.WriteString(fmt.Sprintf(fieldDecl, name, isb.String()) + "\n")
}

//Create the field declaration as well as the Field object. These field objects are used in template generation.
//This ensures the values in json tags will match the database.
//directive name is used for facet.
//TODO: Cleanup this function. It has very mixed logic.
func createField(objectName string, name string, typ string, sb *bytes.Buffer, flag flags, directiveName string, g *Generator) Field {
	var fi Field
	if flag&flagArray != 0 {
		fi.TypeLabel += "[]"
	}
	//time.Time is a special case as it is a struct and therefore has to be a pointer.
	//TODO: Should array be poointer?
	if (flag&flagPointer != 0 /*&& (flag&flagArray == 0)*/) || typ == "time.Time" {
		fi.TypeLabel += "*"
	}
	fi.TypeLabel += typ
	fi.Type = typ

	//Do not capitalize the tag.
	var dbName = objectName + "." + name
	//Ensure for reverse fields the reverse field is used instead.
	if parseState == "dgraph" && (flag&flagReverse > 0 && flag&flagArray > 0) {
		dbName = "~" + typ + "." + directiveName
	} else if flag&flagTwoWay > 0 {
		dbName = "~" + typ + "." + directiveName
	}
	fi.Tag = dbName

	sb.WriteString(strings.Title(name))
	if flag&flagFacet > 0 {
		//if flag&flagReverse > 0 {
		if !(reverseMap[directiveName] != objectName) {
			fi.WrittenTag = "~"
		}
		fi.WrittenTag += reverseMap[directiveName] + "." + directiveName + "|" + name
		fi.Tag = fi.WrittenTag
		//} else {
		//		fi.Tag = objectName + "." + directiveName + "|" + name
		//		fi.WrittenTag = fi.Tag
		//	}
	} else {
		fi.WrittenTag = dbName
	}
	//All tags are omitempty.
	fi.WrittenTag += ",omitempty"

	//TODO: Right now there is an omitempty on non-mandatory fields. This should work I believe.

	sb.WriteString(fmt.Sprintf(lineDeclaration,
		//Ensure it is capitalized for export.
		fi.TypeLabel,
		name,
		fi.WrittenTag))
	fi.Name = strings.Title(name)
	fi.Type = typ
	fi.flags = flag
	return fi
}

//This map remembers the state of reverse edges.
var reverseMap = make(map[string]string)

//Returns the field name relevant for the database. Iterates over non-null & list and marks flags.
//TODO: Do not return pointer!
func iterate(objName string, data *schema.Field, field common.Type, sb *bytes.Buffer, f flags, g *Generator) Field {
	var typ string
	switch a := field.(type) {
	case *common.NonNull:
		return iterate(objName, data, a.OfType, sb, f|flagNotNull, g)
	case *common.List:
		return iterate(objName, data, a.OfType, sb, f|flagArray, g)
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
		/*
			This code is pretty messy but it's important. A reverse edge can be 1-many or many-many.
			For 1-many always use the 1 edgename, but for many-many it is defined by source.
		*/

		//TODO: don't inline interface like this.
		type object interface {
			GetField(string) *schema.Field
			GetName() string
		}
		if dir := data.Directives.Get("hasInverse"); dir != nil {
			f |= flagReverse
			val := dir.Args.MustGet("field")
			var obj object

			if val := g.schema.GetObject(typ); val != nil {
				obj = val
			} else if val := g.schema.GetInterface(typ); val != nil {
				obj = val
			}

			if obj != nil {
				if objField := obj.GetField(val.String()); objField != nil {
					if objField.IsArray() && f&flagArray > 0 {
						//Many-many.
						if arr := data.Directives.Get("source"); arr == nil {
							objName = obj.GetName()
							f |= flagTwoWay
						}
					} else {
						if f&flagArray > 0 {
							//We are the reverse edge.
							objName = obj.GetName()
							f |= flagTwoWay
						} else if objField.IsArray() {
							//We are the 1-edge. Nothing to do here.
						}
					}
					reverseMap[val.String()] = objName
				} else {
					panic("missing field in hasInverse object: " + val.String())
				}
			} else {
				panic("missing field in edge for hasInverse, misspelled? " + val.String())
			}
			dirName = val.String()
		}
		if dir := data.Directives.Get("facet"); dir != nil {
			//Facet value.
			val := dir.Args.MustGet("edge").String()
			f |= flagFacet
			dirName = val
		}
		if dir := data.Directives.Get("lang"); dir != nil {
			f |= flagLang
		}

		var ff = createField(objName, name, typ, sb, f, dirName, g)
		ff.Directives = data.Directives
		ff.Nofield = ff.HasDirective("ignore") != nil
		ff.Nosave = ff.HasDirective("ignore") != nil
		return ff
	}
	return Field{}
}

//The input struct for model template.
type modelStruct struct {
	Name         string
	Fields       []Field
	ScalarFields []Field
	//used for SetType.
	Interfaces []string
	Facets     []string
}

//Executes the model template for all scalar fields.
func modelTemplate(sch *schema.Schema, name string, allScalars []Field, sb *bytes.Buffer, implements []string) {
	templ := getTemplate("Model")
	if templ == nil {
		panic("missing Model template")
	}
	var m = modelStruct{
		Name:   name,
		Fields: allScalars,
	}
	for _, v := range implements {
		m.Interfaces = append(m.Interfaces, v)
	}
	/*for _,v := range m.fields {
		if v.flags&flagFacet > 0 {
			m.Facets = append(m.Facets, strings.ToLower(v.Name))
		}
	}*/
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
				return errors.New("cannot use hasInverse on scalar")
			}
		}
	}
	return nil
}

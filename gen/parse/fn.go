package parse

import (
	"github.com/Vliro/humus/gen/graphql-go/schema"
	"io"
)

type FnCreator struct {
	Fields map[string][]Field
}

func (f FnCreator) Create(i *Generator, w io.Writer) {
	f.processFunctions(i.schema, i.outputs[FunctionFileName], f.Fields)
}

type recurseTemplate struct {
	Fields []Field
	ArrayFields []Field
	Name   string
	Interfaces []string
}

/*
	processFunctions is the entrypoint for declaring all predeclared functions
*/

func (f FnCreator) processFunctions(sch *schema.Schema, writer io.Writer, m map[string][]Field) {
	obj := sch.Objects()
	for _, v := range obj {
		processFieldTemplates(v.Name, writer,m, v.InterfaceNames)
	}
	interf := sch.Interfaces()
	for _, v := range interf {
		processFieldTemplates(v.Name, writer, m, nil)
	}
	makeGlobals(writer)
}
//creates global field functions.
func makeGlobals(writer io.Writer) {
	templ := getTemplate("Field")
	if templ == nil {
		panic("mising field template")
	}
	_ = templ.Execute(writer, nil)
}

//generates using the get.template.
//these functions are for individual fields that are also database objects.
func processFieldTemplates(name string, w io.Writer, m map[string][]Field, intf []string) {
	var output recurseTemplate
	templ := getTemplate("Recurse")
	//asyncTempl := getTemplate("Async")
	if templ == nil {
		panic("missing recurse template")
	}
	for _, v := range m[name] {
		//Fill the data with appropriate data values.
		//Only non-scalar values.
		if v.flags & flagScalar != 0 || v.flags & flagEnum != 0{
			continue
		}
		var data Field
		data.Type = v.Type
		data.Name = v.Name
		data.Parent = name
		data.IsArray = v.flags & flagArray > 0
		data.Tag = v.Tag
		data.TypeLabel = v.TypeLabel
		if data.IsArray {
			output.ArrayFields = append(output.ArrayFields, data)
		} else {
			output.Fields = append(output.Fields, data)
		}
	}
	output.Interfaces = intf
	output.Name = name
	//write to the writer!
	_ = templ.Execute(w, output)
	//_ = asyncTempl.Execute(w, output)
}

package parse

import (
	"github.com/Vliro/mulbase/gen/graphql-go/schema"
	"io"
)

type FnCreator struct {
	Fields map[string][]Field
}

func (f FnCreator) Create(i *Generator, w io.Writer) {
	f.processFunctions(i.schema, i.outputs[FunctionFileName], f.Fields)
}

type GetTemplate struct {
	Fields []Field
	Name   string
}

/*
	processFunctions is the entrypoint for declaring all predeclared functions
*/

func (f FnCreator) processFunctions(sch *schema.Schema, writer io.Writer, m map[string][]Field) {
	//obj := sch.Objects()
	//writeImports(fnImports, writer)
	//for _, v := range obj {
	//	processFieldTemplates(v, writer, m)
	//}
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
func processFieldTemplates(obj *schema.Object, w io.Writer, m map[string][]Field) {
	var output GetTemplate
	templ := getTemplate("Get")
	//asyncTempl := getTemplate("Async")
	if templ == nil {
		panic("missing get template")
	}
	for _, v := range m[obj.Name] {
		//Fill the data with appropriate data values.
		//Only non-scalar values.
		if v.flags & flagScalar != 0 || v.flags & flagEnum != 0{
			continue
		}
		var data Field
		data.Type = v.Type
		data.Name = v.Name
		data.Parent = obj.GetName()
		data.IsArray = v.flags & flagArray > 0
		data.Tag = v.Tag
		data.TypeLabel = v.TypeLabel
		output.Fields = append(output.Fields, data)
	}
	output.Name = obj.Name
	//write to the writer!
	_ = templ.Execute(w, output)
	//_ = asyncTempl.Execute(w, output)
}

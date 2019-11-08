package parse

import (
	"io"
	"mulbase/gen/graphql-go/common"
	"mulbase/gen/graphql-go/schema"
)

type fieldTemplate struct {
	Name   string
	Parent string
	Type   string
}

type GetTemplate struct {
	Fields []fieldTemplate
	Name   string
}

var fnImports = []string{
	"mulbase",
	"context",
}
/*
	processFunctions is the entrypoint for declaring all predeclared functions
*/

func processFunctions(sch *schema.Schema, writer io.Writer) {
	obj := sch.Objects()
	writeImports(fnImports, writer)
	for _, v := range obj {
		processFieldTemplates(v, writer)
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
func processFieldTemplates(obj *schema.Object, w io.Writer) {
	var output GetTemplate
	templ := getTemplate("Get")
	asyncTempl := getTemplate("Async")
	if templ == nil {
		panic("missing get template")
	}
	for _, v := range obj.Fields {
		var typ common.Type = v.Type
		for {
			if val, ok := typ.(*common.NonNull); ok {
				typ = val.OfType
				continue
			}
			if val, ok := typ.(*common.List); ok {
				typ = val.OfType
				continue
			}
			break
		}
		if val, ok := typ.(*schema.Object); ok {
			var data fieldTemplate
			data.Type = val.GetName()
			data.Name = v.GetName()
			data.Parent = obj.GetName()
			output.Fields = append(output.Fields, data)
		}
	}
	output.Name = obj.Name
	//write to the writer!
	_ = templ.Execute(w, output)
	_ = asyncTempl.Execute(w, output)
}

package parse

import (
	"html/template"
	"io"
	"io/ioutil"
	"mulbase/gen/graphql-go/common"
	"mulbase/gen/graphql-go/schema"
)

type GetTemplate struct {
	Name string
	Parent string
	Type string
}

var templates = map[string]string {
	"Get": "templates/get.template",
}

func getTemplate(name string) *template.Template {
	file, ok := templates[name]
	if !ok {
		return nil
	}
	byt, err := ioutil.ReadFile("parse/"+file)
	if err != nil {
		return nil
	}
	templ, err := template.New(name).Parse(string(byt))
	if err != nil {
		return nil
	}
	return templ

}

/*
	processFunctions is the entrypoint for declaring all predeclared functions
*/

func processFunctions(sch *schema.Schema, writer io.Writer) {
	obj := sch.Objects()
	for _,v := range obj {
		createFieldGetter(v, writer)
	}
}

func createFieldGetter(obj *schema.Object, w io.Writer) string {
	templ := getTemplate("Get")
	if templ == nil {
		panic("missing get template")
	}
	for _,v := range obj.Fields {
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
		if val,ok := typ.(*schema.Object); ok {
			var data GetTemplate
			data.Type = val.GetName()
			data.Name = v.GetName()
			data.Parent = obj.GetName()
			templ.Execute(w, data)
		}
	}
	return ""
}

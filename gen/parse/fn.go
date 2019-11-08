package parse

import (
	"github.com/gobuffalo/packr/v2"
	"io"
	"mulbase/gen/graphql-go/common"
	"mulbase/gen/graphql-go/schema"
	"text/template"
)

type fieldTemplate struct {
	Name string
	Parent string
	Type string
}

type GetTemplate struct {
	Fields []fieldTemplate
}

var templates = map[string]string {
	"Get": "get.template",
	"Field": "field.template",
	"Async": "async.template",
}
//The box relevant for embedding assets.
var box = packr.New("templates", "./templates")

func getTemplate(name string) *template.Template {
	file, ok := templates[name]
	if !ok {
		return nil
	}
	str, err := box.FindString(file)
	if err != nil {
		panic(err)
	}
	//byt, err := ioutil.ReadFile("parse/"+file)
	if err != nil {
		return nil
	}
	templ, err := template.New(name).Parse(str)
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
		processFieldTemplates(v, writer)
	}
	makeGlobals(writer)
}

func makeGlobals(writer io.Writer) {
	templ := getTemplate("Field")
	if templ == nil {
		panic("mising field template")
	}
	templ.Execute(writer, nil)
}
//generates using the get.template.
func processFieldTemplates(obj *schema.Object, w io.Writer)  {
	var output GetTemplate
	templ := getTemplate("Get")
	asyncTempl := getTemplate("Async")
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
			var data fieldTemplate
			data.Type = val.GetName()
			data.Name = v.GetName()
			data.Parent = obj.GetName()
			output.Fields = append(output.Fields, data)
		}
	}
	//write to the writer!
	_ = templ.Execute(w, output)
	_ = asyncTempl.Execute(w, output)
}


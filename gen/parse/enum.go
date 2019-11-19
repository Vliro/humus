package parse

import (
	"github.com/Vliro/mulbase/gen/graphql-go/schema"
	"io"
)

type EnumCreator struct{}

func (e EnumCreator) Create(i *Generator, w io.Writer) {
	e.makeEnums(i.enums(), i.outputs[EnumFileName])
}

type EnumValues struct {
	Name string
	Start string
	Fields []string
}

type EnumResult struct {
	Vals []EnumValues
}

func (e EnumCreator) makeEnums(enums []*schema.Enum, output io.Writer) {
	var res EnumResult
	templ := getTemplate("Enum")
	for _,v := range enums {
		var data EnumValues
		for _,name := range v.Values {
			if data.Start == "" {
				data.Start = name.Name
				continue
			}
			data.Fields = append(data.Fields, name.Name)
		}
		data.Name = v.Name
		res.Vals = append(res.Vals, data)
	}
	_ = templ.Execute(output, res)
}

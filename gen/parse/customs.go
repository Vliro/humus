package parse

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"io"
	"strings"
)

//the creator for custom fields.
type CustomCreator struct{}

func (c CustomCreator) Create(i *Generator, w io.Writer) {
	if w == nil {
		return
	}
	val, ok := i.States[CustomsFileName]
	if !ok {
		return
	}
	c.writeCustoms(val.(map[string]map[string]Customs), w, i)
}

type Customs struct {
	Fields []string
}

func (c CustomCreator) writeCustoms(customs map[string]map[string]Customs, output io.Writer, gg *Generator) {
	var sb bytes.Buffer
	for root, v := range customs {
		g := globalFields
		fields, ok := g[root]
		if !ok {
			panic("invalid customs file, missing in global definitions")
		}
		for sub, vv := range v {
			var newFields = make([]Field, len(vv.Fields))
			name := root + strings.Title(sub)
		outer:
			for k, field := range vv.Fields {
				title := strings.Title(field)
				for r, innerField := range fields.AllFields {
					if innerField.Name == title {
						newFields[k] = innerField
						continue outer
					}
					if r == len(fields.AllFields)-1 {
						panic("missing field in definition from custom.toml")
					}
				}
			}
			sb.WriteString("\n// This is a custom field that is defined by custom.toml.\n")
			makeFieldList(name, newFields, &sb, false, gg)
		}
	}
	_, _ = io.Copy(output, &sb)
}

func parseCustoms(input io.Reader) map[string]map[string]Customs {
	m := make(map[string]map[string]Customs)
	_, err := toml.DecodeReader(input, &m)
	if err != nil {
		panic(err)
	}
	return m
}

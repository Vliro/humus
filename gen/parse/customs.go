package parse

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"io"
	"strings"
)

type CustomCreator struct{}

func (c CustomCreator) Create(i *Generator, w io.Writer) {
	if w == nil {
		return
	}
	i.writeImports(defaultImport, w)
	c.writeCustoms(i.States[CustomsFileName].(map[string]map[string]Customs), w)
}

type Customs struct {
	Fields []string
}

func (c CustomCreator) writeCustoms(customs map[string]map[string]Customs, output io.Writer) {
	var sb bytes.Buffer
	for root,v := range customs {
		g := globalFields
		fields,ok := g[root]
		if !ok {
			panic("invalid customs file, missing in global definitions")
		}
		for sub,vv := range v {
			var newFields = make([]Field, len(vv.Fields))
			name := root + strings.Title(sub)
		outer: for k,field := range vv.Fields {
			title := strings.Title(field)
			for r,innerField := range fields.AllFields {
				if innerField.Name == title {
					newFields[k] = innerField
					continue outer
				}
				if r == len(fields.AllFields)-1 {
					panic("missing field in definition from custom.toml")
				}
			}
		}
			makeFieldList(name, newFields, &sb, false)
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


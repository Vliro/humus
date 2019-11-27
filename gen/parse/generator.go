package parse

import (
	"bytes"
	"github.com/Vliro/mulbase/gen/graphql-go"
	"github.com/Vliro/mulbase/gen/graphql-go/schema"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var defaultImport = []string{"github.com/Vliro/mulbase"}

type Creator interface {
	Create(i *Generator, w io.Writer)
}

//Generator handles the state of the entire generation.
type Generator struct {
	config *Config
	//list of file outputs.
	outputs map[string]*bytes.Buffer
	//the active schema.
	schema *schema.Schema
	//metadata
	meta map[string]map[string]*Meta
	//custom fields.
	customs map[string]map[string]Customs
	//To set states from certain function
	States map[string]interface{}
}

func newGenerator(c *Config) *Generator {
	if c == nil {
		panic("nil config")
	}
	gen := new(Generator)
	gen.outputs = make(map[string]*bytes.Buffer)
	gen.config = c
	gen.States = make(map[string]interface{})
	err := gen.prepare()
	if err != nil {
		panic(err)
	}
	return gen
}

func (g *Generator) getObjects() []*schema.Object {
	return g.schema.Objects()
}

func (g *Generator) getInterfaces() []*schema.Interface {
	return g.schema.Interfaces()
}

func (g *Generator) enums() []*schema.Enum {
	return g.schema.Enums()
}

func (g *Generator) Run() {
	var inter Creator
	templ := getTemplate("All")
	for _,v := range g.outputs {
		_ = templ.Execute(v, g.config.Package)
	}
	inter = ModelCreator{}
	g.writeHeader(g.outputs[ModelFileName])
	inter.Create(g, g.outputs[ModelFileName])
	inter = FnCreator{Fields: g.States[ModelFileName].(map[string][]Field)}
	g.writeHeader(g.outputs[FunctionFileName])
	inter.Create(g, g.outputs[FunctionFileName])
	inter = EnumCreator{}
	g.writeHeader(g.outputs[EnumFileName])
	inter.Create(g, g.outputs[EnumFileName])
	inter = CustomCreator{}
	g.writeHeader(g.outputs[CustomsFileName])
	inter.Create(g, g.outputs[CustomsFileName])
	g.finish()
}

func (g *Generator) writeImports(imports []string, out io.Writer) {
	return
	if out == nil {
		return
	}
	var buf bytes.Buffer
	buf.WriteString("import ( \n")
	var newImports = make([]string, len(imports))
	for k,v := range imports {
		newImports[k] = "\"" + v + "\""
	}
	buf.WriteString(strings.Join(newImports, "\n"))
	buf.WriteString("\n)\n")
	_,_ = io.Copy(out, &buf)
}




func (g *Generator) prepare() error {
	/*
		First pass. Generate the gen.
	*/
	config := g.config
	g.outputs[ModelFileName] = new(bytes.Buffer)
	g.outputs[EnumFileName] = new(bytes.Buffer)
	g.outputs[FunctionFileName] = new(bytes.Buffer)

	/*
		Walk the directory.
	*/
	var resultingFile bytes.Buffer
	var hasFile = false
	err := filepath.Walk(config.Input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if !info.IsDir() {
			//This should be safe.
			if fp := filepath.Ext(info.Name()); fp != ".graphql" && fp != ".toml" {
				return nil
			}
			if info.Name() == "dgraph_schema.graphql" {
				return nil
			}
			fd, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}
			/*
				Use graphQL parser for the schema.
			*/
			//
			if fp := filepath.Ext(info.Name()); fp == ".toml" {
				if info.Name() == "meta.toml" {
					g.meta = parseMeta(bytes.NewReader(fd))
					return nil
				}
				if info.Name() == "custom.toml" {
					g.customs = parseCustoms(bytes.NewReader(fd))
					return nil
				}
			}
			hasFile = true
			resultingFile.Write(fd)
			resultingFile.WriteByte('\n')
			//Ensure we add the proper header.
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	//No file found. Don't do anything.
	if !hasFile {
		return errors.New("no files")
	}
	data := getGraphFile("error.graphql")
	if data == nil {
		panic("something went horribly wrong")
	}
	resultingFile.Write(data)
	/*
		Parse the entire schema.
	*/
	sc := graphql.MustParseSchema(resultingFile.String(), nil)
	g.schema = sc.Schema

	customs, err := ioutil.ReadFile(config.Input + "/custom.toml")
	if len(customs) > 0 {
		buf := new(bytes.Buffer)
		g.outputs[CustomsFileName] = buf
		tom := parseCustoms(bytes.NewReader(customs))
		g.States[CustomsFileName] = tom
	}
	return nil
}

func (g *Generator) finish() {
	if g.config.State == "graphql" {
		sch, err := os.Create(g.config.Output + SchemaName)
		if err != nil {
			panic(err)
		}
		defer sch.Close()
	}
	if g.config.State == "dgraph" {
		dgraphSch, err := os.Create(g.config.Output + "/dgraph.txt")
		if err != nil {
			panic(err)
		}
		defer dgraphSch.Close()
		makeSchema(dgraphSch)
	}
	for k,v := range g.outputs {
		if v.Len() > 0 {
			newbuf := goFmt(v.Bytes())
			file, err := os.Create(g.config.Output + k)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(file, bytes.NewReader(newbuf))
			if err != nil {
				panic(err)
			}
			_ =	file.Close()
		}
	}
}

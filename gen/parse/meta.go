package parse

import (
	"github.com/BurntSushi/toml"
	"io"
	"strings"
)

//Meta allows you to specify how the structs are handled in go.
//*.toml files allow you to specify properties that act independent
//of the dgraph database. For instance, you generally don't want password
//fields to be queried per default and/or mutated at times.
//Specifying it as a password excludes it from all default queries.
type Meta struct {
	Password bool
	Facet    string
	Type     string
}

func parseMeta(input io.Reader) map[string]map[string]Meta {
	m := make(map[string]map[string]Meta)
	_, err := toml.DecodeReader(input, &m)
	if err != nil {
		panic(err)
	}
	return m
}

var meta map[string]map[string]Meta

func getMetaValue(name string) *Meta {
	names := strings.Split(name, ".")
	if len(names) != 2 {
		return nil
	}
	if val, ok := meta[names[0]]; ok {
		if md, ok := val[names[1]]; ok {
			return &md
		}
	}
	return nil
}

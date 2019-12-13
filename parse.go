package humus

import (
	"errors"
	"fmt"
	"github.com/Vliro/humus/parse"
	jsoniter "github.com/json-iterator/go"
	"reflect"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

//handleResponse takes the raw input from Dgraph and deserializes into the interfaces
//as provided by inp given the query names. It will use easyjson if available,
//otherwise defaults to standard json.
func handleResponse(res []byte, inp []interface{}, names []string) error {
	i := -1
	//This uses zero memory allocations to traverse the query tree.
	//Since we do not want to deserialize the query root but rather the containing values
	//traversing the query root with zero allocations is a large benefit, making jsonparser
	//a very useful library here.
	//Alternatively you can deserialize into an arbitrary object and use that but it is a lot less efficient.
	return parse.ObjectEach(res, func(key []byte, value []byte, _ parse.ValueType, _ int) error {
		i++
		//Skip empty return values.
		if string(value) == "[]" {
			return nil
		}
		if string(key) != names[i] {
			return errors.New(fmt.Sprintf("mismatch between query name and key: %s : %s", string(key), names[i]))
		}
		return singleResponse(value, inp[i])
	})
}

func singleResponse(value []byte, inp interface{}) error {
	if len(value) == 0 {
		return nil
	}
	val := reflect.TypeOf(inp)
	kind := val.Kind()
	if !(kind == reflect.Ptr) {
		return fmt.Errorf("parse: invalid non-assignable input to singleResponse, got type %s", val.String())
	}
	val = val.Elem()
	kind = val.Kind()
	isArray := kind == reflect.Slice || kind == reflect.Array
	//No need for massive reflect to check if it is an array
	if value[0] == '[' {
		if !isArray {
			value = value[1 : len(value)-1]
		}
	}
	return json.Unmarshal(value, inp)
}

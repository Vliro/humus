package humus

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/mailru/easyjson"
	"reflect"
)

//handleResponse takes the raw input from Dgraph and deserializes into the interfaces
//as provided by inp given the query names. It will use easyjson if available,
//otherwise defaults to standard json.
func handleResponse(res []byte, inp []interface{}, names []string) error {
	i := -1
	//This uses zero memory allocations to traverse the query tree.
	return jsonparser.ObjectEach(res, func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		i++
		//Skip empty values.
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
	val := reflect.TypeOf(inp)
	kind := val.Kind()
	if !(kind == reflect.Ptr || kind == reflect.Interface) {
		return fmt.Errorf("parse: invalid non-assignable input to singleResponse, got type %s", val.String())
	}
	isArray := kind == reflect.Slice || kind == reflect.Array
	if len(value) == 0 {
		return nil
	}
	//No need for massive reflect to check if it is an array
	if value[0] == '[' {
		if !isArray {
			value = value[1 : len(value)-1]
		}
	}
	if val, ok := inp.(easyjson.Unmarshaler); ok {
		return easyjson.Unmarshal(value, val)
	}
	return json.Unmarshal(value, inp)
}

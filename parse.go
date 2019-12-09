package humus

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/mailru/easyjson"
	"reflect"
)

type EmptyResponseErr struct{}

func (e EmptyResponseErr) Error() string {
	return "Empty response from dgraph."
}
/*
//singleResponse parses one response from dgraph into the pointer at inp.
func singleResponse(temp *fastjson.Value, inp interface{}) error {
	r, err := temp.Array()
	if err != nil {
		return err
	}
	if len(r) == 0 {
		return nil
	}
	var b []byte
	val := reflect.ValueOf(inp)
	if !(val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface) {
		return fmt.Errorf("parse: invalid input to singleResponse, got type %s", val.Type().String())
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	isArray := val.Kind() == reflect.Slice || val.Kind() == reflect.Array
	if len(r) == 1 {
		if isArray {
			obj, err := temp.Object()
			if err == nil {
				if obj.Len() == 1 && obj.Get("uid") != nil {
					//No object was actually found, just the uid was found.
					return nil
				}
			}
			b = temp.MarshalTo(b)
			err = json.Unmarshal(b, inp)
			return err
		} else {
			o, err := r[0].Object()
			if err != nil {
				return err
			}
			if o.Len() == 1 && o.Get("uid") != nil {
				//No object was actually found, just the uid was found.
				return nil
			}
			b = o.MarshalTo(b)
			err = json.Unmarshal(b, inp)
			return err
		}
	} else {
		if !isArray {
			o, err := r[0].Object()
			if err != nil {
				return err
			}
			if o.Len() == 1 && o.Get("uid") != nil {
				//No object was actually found, just the uid was found.
				return nil
			}
			byt := r[0].MarshalTo(nil)
			err = json.Unmarshal(byt, inp)
			return err
		} else {
			//val := temp.MarshalTo(nil)
			byt := temp.MarshalTo(nil)
			if err != nil {
				return err
			}
			err = json.Unmarshal(byt, inp)
			return err
		}
	}
}

//HandleResponse handles the input from a query.
func HandleResponse(res []byte, inp []interface{}, names ...string) error {
	//Use a fastjson parser to traverse it initially.
	var parse fastjson.Parser
	v, err := parse.ParseBytes(res)
	if err != nil {
		return err
	}
	d, err := v.Object()
	if err != nil {
		return err
	}
	if d.Len() != len(inp) {
		return errInvalidLength
	}
	if len(names) != 0 {
		for k, v := range inp {
			err = singleResponse(d.Get(names[k]), v)
			if err != nil {
				return errParsing
			}
		}
		return nil
	}
	//Do we have a single query or multiple?
	if q := d.Get("q"); q != nil {
		err = singleResponse(q, inp[0])
		return err
	}
	for k, v := range inp {
		err = singleResponse(d.Get("q"+strconv.Itoa(k)), v)
		if err != nil {
			return errParsing
		}
	}
	return nil
}
*/
func HandleResponseFast(res []byte, inp []interface{}, names []string) error {
	i := -1
	return jsonparser.ObjectEach(res, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		i++
		if string(key) != names[i] {
			return errors.New(fmt.Sprintf("mismatch between query name and key: %s : %s", string(key), names[i]))
		}
		return singleResponseFast(value, inp[i])
	})
}

func singleResponseFast(value []byte, inp interface{}) error {
	val := reflect.TypeOf(inp)
	kind := val.Kind()
	if !(kind == reflect.Ptr || kind == reflect.Interface) {
		return fmt.Errorf("parse: invalid input to singleResponse, got type %s", val.String())
	}
	isArray := kind == reflect.Slice || kind == reflect.Array
	if len(value) == 0 {
		return nil
	}
	//No need for massive reflect to check if it is an array
	if value[0] == '[' {
		if !isArray {
			value = value[1:len(value)-1]
		}
	}
	if val, ok := inp.(easyjson.Unmarshaler); ok {
		return easyjson.Unmarshal(value, val)
	}
	return json.Unmarshal(value, inp)
}
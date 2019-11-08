package mulbase

import (
	"PR2Server/core/logger"
	"errors"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/valyala/fastjson"
)

type EmptyResponseErr struct{}

func (e EmptyResponseErr) Error() string {
	return "Empty response from dgraph."
}

//json is a variable in this package
func GetResponse(res []byte, inp interface{}) {
	var f map[string]interface{}

	err := json.Unmarshal(res, &f)
	if err != nil {
		return
	}
	for _, v := range f {
		s := v.([]interface{})
		if len(s) > 0 {
			config := &mapstructure.DecoderConfig{Metadata: nil, TagName: "json", Result: &inp}
			decoder, err := mapstructure.NewDecoder(config)
			if err != nil {
				panic(err)
			}
			err = decoder.Decode(s)
			if err != nil {
				panic(err)
			}
			return
		}
	}
	return
}
/*
	func HandleResponseArray(res []byte, params []interface{}) error {
	p := fastjson.Parser{}
	val, err := p.ParseBytes(res)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(params); i++ {
		obj := val.Get("q" + strconv.Itoa(i))
		if obj != nil {
			err = singleResponse(obj, params[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
*/
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
		logger.LogWarn("invalid input to singleResponse.")
		return errors.New("parse: invalid input to singleResponse")
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
			b = o.MarshalTo(b)
			err = json.Unmarshal(b, inp)
			return err
		} else {
			val := temp.MarshalTo(nil)
			err = json.Unmarshal(val, inp)
			return err
		}
	}
	return nil
}
//HandleResponse handles the input from a query..
func HandleResponse(res []byte, inp []interface{}) error {
	p := fastjson.Parser{}
	v, err := p.ParseBytes(res)
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
	for k,v := range inp {
		err = singleResponse(d.Get("q" + strconv.Itoa(k)), v)
		if err != nil {
			return errParsing
		}
	}
	return nil
}

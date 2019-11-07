package mulbase

import (
	"encoding/json"
	"unsafe"

	"github.com/mitchellh/mapstructure"
)

//As string is immutable this is mostly safe as long as we dont change.
func bytesToStringUnsafe(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func Serialize(d interface{}) string {
	bytes, err := json.Marshal(d)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func Deserialize(s string, class interface{}) error {
	bytes := []byte(s)

	err := json.Unmarshal(bytes, class)

	return err
}

func DeserializeByte(b []byte, class interface{}) error {
	err := json.Unmarshal(b, class)

	return err
}
//StructToMap creates a map from struct.
func StructToMap(input interface{}, output interface{}) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, output)
	return err
}

//DeserializeFromMap takes a map and class and parses the struct into map.
func DeserializeFromMap(m map[string]interface{}, class interface{}) {
	config := &mapstructure.DecoderConfig{Metadata: nil, TagName: "json", Result: class}
	decoder, err := mapstructure.NewDecoder(config)

	if err != nil {
		panic(err)
	}
	err = decoder.Decode(m)
	if err != nil {
		panic(err)
	}
	return
}

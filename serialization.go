package mulbase

import (
	jsoniter "github.com/json-iterator/go"
	"io"
	"unsafe"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

//As string is immutable this is mostly safe as long as we dont change any value in b!
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

func ObjectToJson(obj interface{}, w io.Writer) error {
	return json.NewEncoder(w).Encode(obj)
}

func JsonToObject(obj interface{}, r io.Reader) error {
	return json.NewDecoder(r).Decode(obj)
}
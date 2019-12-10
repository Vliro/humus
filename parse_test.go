package humus

import (
	"testing"
)

var res = []byte(`{
    "q0": [
      {
        "Error.errorType": "test",
        "Error.time": "2019-12-10T13:47:47.967597476+01:00",
        "Error.message": "testagain"
      }
    ]
  }`)

func TestDeserialize(t *testing.T) {
	var err dbError
	er := handleResponse(res, []interface{}{&err}, []string{"q0"})
	if er != nil {
		t.Fail()
		return
	}
	if err.Message != "testagain" || err.ErrorType != "test" {
		t.Fail()
	}
}
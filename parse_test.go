package humus

import (
	"testing"
)

var testErrorValue = []byte(`{
    "q0": [
      {
        "Error.errorType": "test",
        "Error.time": "2019-12-10T13:47:47.967597476+01:00",
        "Error.message": "testagain"
      }
    ]
  }`)

func TestDeserialize(t *testing.T) {
	var res dbError
	err := handleResponse(testErrorValue, []interface{}{&res}, []string{"q0"})
	if err != nil {
		t.Fail()
		return
	}
	if res.Message != "testagain" || res.ErrorType != "test" {
		t.Fail()
	}
}
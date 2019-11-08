package mulbase

import (
	"errors"
	"fmt"
)

var errInvalidType = errors.New("invalid query supplied")
var errInvalidLength = errors.New("invalid number of inputs")
var errParsing = errors.New("error parsing input")
var errTransaction = errors.New("invalid transaction")
var ErrUID = errors.New("missing UID")

//Call this on a top level error.
func Error(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("mulbase: %v", err.Error())
}

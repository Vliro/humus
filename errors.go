package humus

import (
	"errors"
	errors2 "github.com/pkg/errors"
)

var errInvalidType = errors.New("invalid query supplied")
var errInvalidLength = errors.New("invalid number of inputs specified to deserialize")
var errParsing = errors.New("error parsing input")
var errTransaction = errors.New("invalid transaction")
var ErrUID = errors.New("missing UID")

var errMissingFunction = errors.New("missing function")
var errMissingVariables = errors.New("missing variables in function")

//Call this on a top level error.
func Error(err error) error {
	if err == nil {
		return nil
	}

	return errors2.Wrap(err, "mulbase")
}

var fErrNil = "nil check failed for %s"

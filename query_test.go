package mulbase

import (
	"context"
	"testing"
)

func TestQuery(t *testing.T) {
	q := GeneratedQuery{}
	txn := new(Txn)
	txn.RunQuery(context.Background(), &q, nil)
}
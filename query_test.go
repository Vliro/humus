package mulbase

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

var alpha, zero *exec.Cmd

func TestMain(m *testing.M) {
	d := setup()
	makeSchema(d)
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func makeSchema(d *DB) {
	q := StaticQuery{}
	q.Query = "schema{}"
	var m = make(map[string]interface{})
	d.NewTxn(true).RunQuery(context.Background(), q, &m)
	print(m)
}

func TestQuery(t *testing.T) {

}

func setup() *DB {
	dz := exec.Command("dgraph zero")
	da := exec.Command("dgraph alpha")
	dz.Run()
	da.Run()
	zero = dz
	alpha = da

	return Init("localhost", 9080, false)
}

func shutdown() {
	alpha.Process.Kill()
	zero.Process.Kill()
}
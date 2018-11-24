package wpgx_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
)

// psql:
//
// create database wpgx;
// create user wpgx with password 'wpgx';
// grant all on DATABASE wpgx to wpgx;
//
// connStr:
// postgres://wpgx:wpgx@127.0.0.1:5432/wpgx?sslmode=disable
const connFile = "test.connstr"

var connStr string = ""
var reserve string = filepath.Join("testdata", "reserve")

func TestMain(m *testing.M) {
	var err error

	if err = setup(); err != nil {
		log.Fatalf("%+v", err)
	}

	code := m.Run()

	if err = teardown(); err != nil {
		log.Fatalf("%+v", err)
	}

	os.Exit(code)
}

func setup() (err error) {
	const emsg = "setup tests"

	if connStr == "" {
		var data []byte

		if data, err = ioutil.ReadFile(connFile); err != nil {
			return errors.Wrap(err, emsg)
		}
		connStr = string(data)
	}

	if err = os.MkdirAll(reserve, 0755); err != nil {
		return errors.Wrap(err, emsg)
	}

	return
}

func teardown() (err error) {
	const emsg = "teardown tests"

	if err = os.RemoveAll(reserve); err != nil {
		return errors.Wrap(err, emsg)
	}

	return
}

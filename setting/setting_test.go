package setting

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"qing/go-helper/error"
	testingX "qing/go-helper/testing"
	"testing"
)

func Test_initFromFile(t *testing.T) {
	// situation 1: not support config file format
	path := "exist.jjj"
	err := initFromFile(path, nil)
	wantedErr := errPkg.Fail("not supported file format", errPkg.Fields{
		"file": path,
		"exts": []string{".json"},
	})
	assert.Equal(t, wantedErr, err, "situation 1: not support config file format")

	// situaton 2: open file fail
	path = "not-exist.json"
	err = initFromFile(path, nil)
	wantedErr = func(file string) *errPkg.Err {
		_, err := os.Open(file)
		return errPkg.FailBy(err, "open config file fail", errPkg.Fields{"file": file})
	}(path)
	assert.Equal(t, wantedErr, err, "situaton 2: open file fail")

	// situation 3: unmarshal fail

	v1 := make([]string, 0)
	mockJson := `{"omg":true, "size":10}`
	path = "1.json"
	f := testingX.MockFile(path, mockJson)
	err = initFromFile(path, &v1)
	wantedErr = func(jsonStr string) *errPkg.Err {
		err := json.Unmarshal([]byte(jsonStr), &v1)
		return errPkg.FailBy(err, "unmarshal config file's content bytes to confObj fail",
			errPkg.Fields{"file": path})
	}(mockJson)
	assert.Equal(t, wantedErr, err, "situation 3: unmarshal fail")
	f.Remove()

	// ok
	type beauty struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	v2 := new(beauty)
	mockJson = `{"name":"xiaolonglv", "age":100}`
	path = "2.json"
	f = testingX.MockFile(path, mockJson)
	err = initFromFile(path, &v2)
	assert.Equal(t, nil, err, "ok")
	assert.Equal(t, beauty{"xiaolonglv", 100}, *v2)
	f.Remove()
}

func Test_InitFromFile(t *testing.T) {
	expected := new(OneConf)
	expected.A = 100
	expected.B = "200"

	path, _ := expected.FromFile()
	mockJson := `{"x":{"y":{"z":{"a":100,"b":"200"}}}}`
	f := testingX.MockFile(path, mockJson)
	defer f.Remove()

	actual := new(OneConf)
	if err := InitFromFile(actual); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected, actual)
}

type OneConf struct {
	A int    `json:"a"`
	B string `json:"b"`
}

func (conf OneConf) FromFile() (string, []string) {
	return "conf.json", []string{"x", "y", "z"}
}

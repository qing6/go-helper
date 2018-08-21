package errPkg

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"errors"
)

func TestErr_Error(t *testing.T) {
	err := FailBy(errors.New("this is cause"), "this is error", Fields{"0":1})
	expected := "Error: this is error\tinfo={0=1,\b}\tcause=this is cause"
	actual := err.Error()
	assert.Equal(t, expected, actual)
	//t.Log(expected)
}

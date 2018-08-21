package errPkg

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

type Err struct {
	CreatedAt time.Time  `json:"createdAt"`
	Msg       string     `json:"msg"`
	Fields    Fields     `json:"fields"`
	Cause     error      `json:"cause"`
	StackInfo *StackInfo `json:"stackInfo"`
}

type Fields map[string]interface{}

func Fail(msg string, fields Fields) *Err {
	return &Err{
		CreatedAt: time.Now(),
		Msg:       msg,
		Fields:    fields,
	}
}

func FailBy(err error, msg string, fields Fields) *Err {
	e := Fail(msg, fields)
	e.Cause = err
	return e
}

func (err *Err) Error() string {
	bf := bufferPool.Get().(*bytes.Buffer)
	bf.Reset()
	defer bufferPool.Put(bf)
	bf.WriteString("Error: ")
	bf.WriteString(err.Msg)
	bf.WriteString("\tinfo={")
	for k, v := range err.Fields {
		bf.WriteString(k)
		bf.WriteString("=")
		bf.WriteString(fmt.Sprintf("%+v", v))
		bf.WriteString(",")
	}
	bf.WriteString("\b}")
	if err.Cause != nil {
		bf.WriteString("\tcause=")
		bf.WriteString(err.Cause.Error())
	}
	if err.StackInfo != nil {
		bf.WriteString("\tstack=")
		bf.WriteString(err.StackInfo.String())
	}
	return bf.String()
}

func (err *Err) SetField(k string, v interface{}) *Err {
	if err.Fields == nil {
		err.Fields = make(Fields)
	}
	err.Fields[k] = v
	return err
}

func (err *Err) Where(pkg, function string) *Err {
	info := new(StackInfo)
	info.Package = pkg
	info.Function = function
	err.StackInfo = info
	return err
}

type StackInfo struct {
	Package  string `json:"package"`
	Function string `json:"function"`
	Code     string `json:"code"`
}

func (info StackInfo) String() string {
	bf := bufferPool.Get().(*bytes.Buffer)
	bf.Reset()
	defer bufferPool.Put(bf)
	bf.WriteString(info.Package)
	bf.WriteString(" ")
	bf.WriteString(info.Function)
	if len(info.Code) != 0 {
		bf.WriteString(" ")
		bf.WriteString(info.Code)
	}
	return bf.String()
}

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func GetBufferPool() *sync.Pool {
	return bufferPool
}

func Wrap(err error, msg string, fields Fields) error {
	if err != nil {
		return FailBy(err, msg, fields)
	}
	return nil
}
package errPkg

import (
	"bytes"
	"fmt"
	"time"
	"qing/go-helper/common"
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
	bf := common.BytesBufferPool.Get().(*bytes.Buffer)
	bf.Reset()
	defer common.BytesBufferPool.Put(bf)
	bf.WriteString("Error: ")
	bf.WriteString(err.CreatedAt.Format(time.RFC3339Nano))
	bf.WriteString(" ")
	bf.WriteString(err.Msg)
	bf.WriteString(" info= {")
	for k, v := range err.Fields {
		bf.WriteString(k)
		bf.WriteString("=")
		bf.WriteString(fmt.Sprintf("%+v", v))
		bf.WriteString(",")
	}
	bf.WriteString("\b}")
	if err.Cause != nil {
		bf.WriteString(" cause= `")
		bf.WriteString(err.Cause.Error())
		bf.WriteString("`")
	}
	if err.StackInfo != nil {
		bf.WriteString(" stack= ")
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

func GetCause(err error) error {
	known, ok := err.(*Err)
	if ok && known.Cause != nil{
		return GetCause(known.Cause)
	}
	return err
}

type StackInfo struct {
	Package  string `json:"package"`
	Function string `json:"function"`
	Code     string `json:"code"`
}

func (info StackInfo) String() string {
	bf := common.BytesBufferPool.Get().(*bytes.Buffer)
	bf.Reset()
	defer common.BytesBufferPool.Put(bf)
	bf.WriteString(info.Package)
	bf.WriteString(" ")
	bf.WriteString(info.Function)
	if len(info.Code) != 0 {
		bf.WriteString(" ")
		bf.WriteString(info.Code)
	}
	return bf.String()
}

func Wrap(err error, msg string, fields Fields) error {
	if err != nil {
		return FailBy(err, msg, fields)
	}
	return nil
}

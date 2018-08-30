// 一些包的使用, 可能需要 config parameters, 通常我会将它们定义为一个 struct
// 比如:
//   使用 logger, 可能需要: Level: 日志级别;  LogDir: 输出日志目录
//   type Conf struct {
//     Level Level
//     Dir   string
//   }
//
//	 使用 db, 可能需要: url
//   type Conf struct {
//     URL string
//   }
//
//   ...
//
// 那么, 如何定义 confObj 的赋值方式?
// source types / 源 优先级:
//   runtime args / 运行时参数
//   config file / 配置文件
//   os environment / 环境变量
//   Apollo... / Apollo 那样的配置中心
//
// 每一个源, 我都提供了一个接口. 只要 confObj 实现这个接口, 该 confObj 将被视为可以通过相应的源的方式来获取赋值
// 一个 confObj 可以实现多个接口, 而源的优先级见上
//
// 一些源的功能实现, 使用了 reflect , 所以建议该包的使用场景位于 所在包的 init() 中
package setting

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"qing/go-helper/error"
	"reflect"
	"strings"
)

func Init(v interface{}) (err error) {
	defer func() {
		if err != nil {
			return
		}
		if needCheck, ok := v.(CanChecked); ok {
			if err = needCheck.Access(); err != nil {
				err = errPkg.FailBy(err, "do check by Access() fail.", nil)
			}
		}
	}()

	notImp := errors.New("not impl")
	var unmarshalErr *errPkg.Err
	wrapUnmarshal := func(source string, optErr error) bool {
		if optErr == nil {
			unmarshalErr = nil
			return true
		}
		if unmarshalErr == nil {
			unmarshalErr = errPkg.Fail("do unmarshal fail.", nil)
		}
		unmarshalErr.SetField(source, source)
		return false
	}

	// cry :(: connot fallthrough in type switch...
	if _, ok := v.(FromOsArgs); ok && wrapUnmarshal("from os-args", notImp) {
		return unmarshalErr
	}

	if confObj, ok := v.(FromFile); ok && wrapUnmarshal("from file", InitFromFile(confObj)) {
		return unmarshalErr
	}

	if _, ok := v.(FromOsEnvs); ok && wrapUnmarshal("from os-envs", notImp) {
		return unmarshalErr
	}

	if _, ok := v.(FromApollo); ok && wrapUnmarshal("from apollo", notImp) {
		return unmarshalErr
	}

	return nil
}

// TODO 实现从运行参数中拉取配置, 参考或使用 github.com/jessevdk/go-flags (https://github.com/jessevdk/go-flags)
// WARN 注意每种 soruce 的实现, 出错不能改变 v 默认值
type FromOsArgs interface {
}

// 从配置文件拉取配置
// WARN 注意每种 soruce 的实现, 出错不能改变 v 默认值
type FromFile interface {

	// path: 配置文件路径
	//   将会根据文件的扩展名来判断解析的方法, 目前支持 json
	// sections: 配置节点名
	//   当多个 struct 配置的 path 相同, 那么建议为每个 struct 配置一个 section
	//   例如:
	//   package log
	//   type Conf struct {
	//     Level Level
	//     Layouts []Layout
	//   }
	//   func (conf *Conf) FromFile() (string, []string) {
	//     return "conf.json", []string{"log"}
	//   }
	//
	//   package db
	//   type Conf struct {
	//     URL string
	//   }
	//   func (conf *Conf) FromFile() (string, []string) {
	//	    return "conf.json", []string{"db"}
	//	 }
	//   那么, 配置文件 conf.json, 应该是这样:
	//
	//   {
	//     "log": {
	//       "level":"DEBUG",
	//       "layouts": ["log.json", "SYS"],
	//     }
	//     "db": {
	//       "url":"......."
	//     }
	//   }
	FromFile() (path string, sections []string)
}

func InitFromFile(v FromFile) error {
	path, sections := v.FromFile()

	if len(sections) == 0 {
		return initFromFile(path, v)
	}

	curSection := sections[len(sections)-1]
	trueVType := reflect.StructOf([]reflect.StructField{
		{
			Name: strings.ToUpper(curSection),
			Type: reflect.TypeOf(v).Elem(),
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, sections[len(sections)-1])),
		},
	})
	for i := len(sections) - 1; i > 0; i-- {
		curSection = sections[i-1]
		trueVType = reflect.StructOf([]reflect.StructField{
			{
				Name: strings.ToUpper(curSection),
				Type: trueVType,
				Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, curSection)),
			},
		})
	}

	trueV := reflect.New(trueVType)
	if err := initFromFile(path, trueV.Interface()); err != nil {
		return err
	}

	vVal := trueV.Elem()
	for range sections {
		vVal = vVal.Field(0)
	}

	var setErr error
	defer func() {
		if panicO := recover(); panicO != nil {
			setErr = errPkg.Fail("using reflect do set fail.", errPkg.Fields{"panic": fmt.Sprint(panicO)})
		}
	}()
	reflect.ValueOf(v).Elem().Set(vVal)
	return setErr
}

func initFromFile(path string, v interface{}) error {
	// 不同格式的文件解析成对象的方法
	// TODO X support more file type, like YAML, XML...
	type fromFile func(f *os.File, v interface{}) error
	fromFileRecords := map[string]fromFile{
		".json": func(f *os.File, v interface{}) error {
			return json.NewDecoder(f).Decode(v)
		},
	}

	parse := fromFileRecords[filepath.Ext(path)]
	if parse == nil {
		return errPkg.Fail("file format is not supported.", errPkg.Fields{
			"file": path,
			"supported extensions": func() []string {
				exts := make([]string, 0)
				for k, _ := range fromFileRecords {
					exts = append(exts, k)
				}
				return exts
			}(),
		})
	}

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return errPkg.FailBy(err, "open config file fail.", errPkg.Fields{"file": path})
	}

	if err = parse(f, v); err != nil {
		return errPkg.FailBy(err, "unmarshal config file's content bytes to confObj fail.",
			errPkg.Fields{"file": path})
	}

	return nil
}

// TODO 实现从环境变量中拉取配置
// WARN 注意每种 soruce 的实现, 出错不能改变 v 默认值
type FromOsEnvs interface {
}

// TODO 实现从 Apollo 拉取配置
// WARN 注意每种 soruce 的实现, 出错不能改变 v 默认值
type FromApollo interface {
}

// confObj 实现该接口表示希望被校验, 具体校验逻辑在 Access() 中, 当然一些默认值的设置也可以在该方法中
type CanChecked interface {
	Access() error
}

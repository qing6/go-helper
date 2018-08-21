package testing

import (
	"os"
	"flag"
)

// Problem 18082001
// 写测试的时候, 出现一个蛋痛的问题
// 在一个项目中, 经常会有 subPkg: conf, 用来存放一些需要配置的信息, 而这些信息往往是从配置文件读取的, 而且功能实现在 init() 中
// 这样导致在测试 conf 包下的方法, 需要提前有个配置文件, 然而我不想手动建个文件, 而是用代码来表达 (代码发布流程应该先运行测试代码)
type CanRemove interface{
	Remove()
}

func MockFile(path string, content string) CanRemove {
	var isNew bool
	if _, err := os.Stat(path); os.IsNotExist(err) {
		isNew = true
	}

	//log.Printf("create file<:%s> if not exist", path)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	defer f.Close()
	if err != nil {
		panic(err.Error())
	}

	//log.Print("write mock content")
	if _, err = f.WriteString(content); err != nil {
		panic(err.Error())
	}

	return toRemoveFile{path: path, isNew: isNew}
}

type toRemoveFile struct {
	path string
	isNew bool
}

func (o toRemoveFile) Remove() {
	if !o.isNew {
		return
	}

	//log.Printf("remove file<:%s>", o)
	if err := os.Remove(o.path); err != nil {
		panic(err.Error())
	}
}

// Problem 18082001-01
// 当某个 pkg 调用 conf, 我无法实现 先于 conf init() 执行一些代码, 如 MockFile...
// Solution
// https://stackoverflow.com/questions/14249217/how-do-i-know-im-running-within-go-test
// Example
//
// func init() {
//   conf = new(Conf)
//
//	  if testing.IsInTesting() {
//		  f := testing.MockFile("conf.json", `{
//			  "accessKey":"12345",
//			  "secretKey":"67890",
//			  "persistDir":"root_dir"
//		  }`)
//        defer f.Remove()
//	  }
//
//	  if err := setting.InitFromFile(conf); err != nil {
//		  panic(err.Error())
//	  }
//  }
func IsInTesting() bool {
	return flag.Lookup("test.v") != nil
}
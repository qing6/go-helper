package testing

import (
	"os"
	"flag"
	"qing/go-helper/error"
	"time"
	"path/filepath"
	"io"
	"strings"
	"qing/go-helper/common"
	"bytes"
)

// Problem 18082001
// 写测试的时候, 出现一个蛋痛的问题
// 在一个项目中, 经常会有 subPkg: conf, 用来存放一些需要配置的信息, 而这些信息往往是从配置文件读取的, 而且功能实现在 init() 中
// 这样导致在测试 conf 包下的方法, 需要提前有个配置文件, 然而我不想手动建个文件, 而是用代码来表达 (代码发布流程应该先运行测试代码)
type CanRemove interface{
	Remove() error
}

type NilCanRemove struct{}

func (canRemove NilCanRemove) Remove() error {
	return nil
}

func MockFile(path string, content string) CanRemove {
	if !IsInTesting() {
		return new(NilCanRemove)
	}

	result := new(mockFileNote)
	result.file = path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fileExists := new(fileExists)
		ext := filepath.Ext(path)
		bf := common.BytesBufferPool.Get().(*bytes.Buffer)
		defer common.BytesBufferPool.Put(bf)
		moveTo := func(originPath string) string {
			bf.Reset()
			bf.WriteString(strings.TrimRight(originPath, ext))
			bf.WriteString("_0")
			bf.WriteString(ext)
			return bf.String()
		}

		fileExists.movedName = moveTo(path)
		for {
			if stat, errX := os.Stat(fileExists.movedName); os.IsNotExist(err) {
				fileExists.modTime = stat.ModTime()
				fileExists.mode = stat.Mode()
				result.fileExists = fileExists
				return result
			} else if errX != nil {
				result.err = errPkg.FailBy(errX, "cal new file path, check stat fail", nil)
				return result
			}
			fileExists.movedName = moveTo(fileExists.movedName)
		}

	} else if err != nil {
		result.err = errPkg.FailBy(err, "check path stat fail", nil)
		return result
	}

	fileNotExists := new(fileNotExists)
	fileNotExists.createdDirs = make([]string, 0)
	parts := strings.Split(filepath.ToSlash(filepath.Dir(path)), "/")



















	bf := common.BytesBufferPool.Get().(*bytes.Buffer)
	defer common.BytesBufferPool.Put(bf)
	bf.Reset()
	for i := 0; i <{
		curDir := bf.String()
		if _, err := os.Stat(curDir); os.IsNotExist(err) {
			fileNotExists.createdDirs = append(fileNotExists.createdDirs, bf.String())
			if err := os.Mkdir(bf.String()) {

			}



		} else if err != nil {
			result.err = errPkg.FailBy(err, "check dir stat fail", nil)
			return result
		}
	}













	for {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.mk






		} else if err != nil {
			result.err = errPkg.FailBy(err, "cal dir path, check if it exists fail", nil)
			return result
		}

	}























	var isNew bool


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

type mockFileNote struct {
	err error
	file string
	fileExists *fileExists
	fileNotExists *fileNotExists
}

type fileExists struct {
	movedName string
	mode os.FileMode
	modTime time.Time
}

func (note fileExists) backout(file string) error {
	movedFile := filepath.Join(filepath.Dir(file), note.movedName)
	mf, err := os.OpenFile(movedFile, os.O_RDONLY, note.mode)
	if err != nil {
		return errPkg.FailBy(err, "open moved file by read_only fail", nil)
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, note.mode)
	if err != nil {
		return errPkg.FailBy(err, "re-create file fail", nil)
	}

	if _, err := io.Copy(f, mf); err != nil {
		f.Close()
		mf.Close()
		os.Remove(file)
		return errPkg.FailBy(err, "exec copy fail", nil)
	}

	f.Close()
	mf.Close()
	os.Remove(movedFile)
	os.Chtimes(movedFile, note.modTime, note.modTime)
	return nil
}

type fileNotExists struct {
	createdDirs []string
}

func (note fileNotExists) backout() error {
	for i:=len(note.createdDirs); i>0; i-- {
		if err := os.Remove(note.createdDirs[i-1]); err != nil {
			return errPkg.FailBy(err, "remove created dir fail", nil)
		}
	}
	return nil
}









func (note mockFileNote) Remove() error {
	if err := note.err; err != nil {
		return err
	}

	if err := os.Remove(note.f)








	if len(notes) == 0 {
		return nil
	}
	for _, note := range notes {
		if err := os.Remove(note); err != nil {
			return errPkg.FailBy(err, "remove note fail", errPkg.Fields{"path":note})
		}
	}
	return nil
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
func IsInTesting() bool {
	return flag.Lookup("test.v") != nil
}
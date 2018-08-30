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
// 写给一个测试函数的时候, 出现一个蛋痛的问题：
//
// 被测试函数的所在包 A , 引用了另外一个包 conf , 该包下的一些变量的赋值逻辑在 init() 中, 这些变量使用了 go-helper/setting 里设计
// 的功能, 即, impl setting.FromFile. 所以, 我需要一个配置文件. (如果文件不存在, setting.Init 会返回一个error, 而我通常的处理,
// 一般是panic, 以表示配置是必需的)
//
// 然而, 写一个测试需要先创建一个配置文件, 这让我不开心. 更不开心的是, 我要手动删除它. (impl setting.FromFile的时候, 配置文件的路径
// 通常是相对路径)
//
// 所以我希望的调用逻辑是这样
//
// type Conf struct {
//   FieldA string
//   ...
// }
//
// func(conf *Conf) FromFile() (string, []string) {
//    return "conf.json", nil
// }
//
// var conf *Conf
//
// func init() {
//   conf = new(Conf)
//
//   f : = testing.MockFile(`conf.json`, `...`)
//   defer f.Remove()
//
//   if setting.Init(conf); err != nil {
//      panic(err.Error)
//   }
// }
//
// 所以, MockFile 执行逻辑:
// 0  判断是否是 go test 场景
// 1  判断 file 是否存在
//   1.0  file 存在
//     1.0.0  为已经存在的 file 找到合适的新的 name
//     1.0.1  os.Rename
//   1.1  file 不存在
//     1.1.0  创建相关的 dir 并记录
//     1.1.1  创建文件
//
func MockFile(path string, content string) CanRemove {
	if !IsInTesting() {
		return new(doNothing)
	}

	return newMockingFileNote(path).Write(content)
}

type CanRemove interface{
	Remove() error
}

type doNothing struct{}

func (canRemove doNothing) Remove() error {
	return nil
}

type mockingFileNote struct {
	processErr error
	file string
	opt mockingFileOpt
}

type mockingFileOpt interface {
	forward(file string) error
	back(file string) error
	name() string
}

func newMockingFileNote(path string) *mockingFileNote {
	result := new(mockingFileNote)
	result.file = path
	if stat, err := os.Stat(path); os.IsNotExist(err) {
		result.opt, result.processErr = mockNonExistentFile(path)
	} else if err != nil {
		result.processErr = err
		return result
	} else {
		result.opt, result.processErr = mockExistentFile(path, stat)
	}
	return result
}

func (note *mockingFileNote) Write(content string) *mockingFileNote {
	if note.processErr == nil {
		note.processErr = errPkg.Wrap(note.opt.forward(note.file),
			"mocking-file-note's opt exec forward fail", errPkg.Fields{"opt":note.opt.name()})
	}

	if note.processErr == nil {
		f, err := os.OpenFile(note.file, os.O_WRONLY, os.ModePerm)
		defer f.Close()
		if err != nil {
			note.processErr = errPkg.FailBy(err, "mocking-file-note open file to write fail.", nil)
		}
		if _, err = f.WriteString(content); err != nil {
			note.processErr = errPkg.FailBy(err, "mocking-file-note write content fail.", nil)
		}
	}

	return note
}

func (note mockingFileNote) Remove() error {
	if note.processErr != nil {
		return note.processErr
	}
	return errPkg.Wrap(note.opt.back(note.file),
		"mocking-file-note's opt exec back fail.", errPkg.Fields{"opt":note.opt.name()})
}

type mockingExistentFileNote struct {
	newFile string
	mode os.FileMode
	modTime time.Time
}

func mockExistentFile(file string, fStat os.FileInfo) (*mockingExistentFileNote, error) {
	ext := filepath.Ext(file)
	bf := common.BytesBufferPool.Get().(*bytes.Buffer)
	defer common.BytesBufferPool.Put(bf)

	gainNewFile := func(originFile string) string {
		bf.Reset()
		bf.WriteString(strings.TrimRight(originFile, ext))
		bf.WriteString("_0")
		bf.WriteString(ext)
		return bf.String()
	}

	newFile := gainNewFile(file)
	for {
		if _, err := os.Stat(newFile); os.IsNotExist(err) {
			note := new(mockingExistentFileNote)
			note.newFile = newFile
			note.modTime = fStat.ModTime()
			note.mode = fStat.Mode()
			return note, nil
		} else if err != nil {
			err = errPkg.FailBy(err, "gain new file path, check if it is available fail.", nil)
			return nil, err
		}
		newFile = gainNewFile(newFile)
	}
}

func (note mockingExistentFileNote) forward(file string) error {
	nf, err := os.OpenFile(note.newFile, os.O_CREATE|os.O_WRONLY, note.mode)
	defer nf.Close()
	if err != nil {
		return errPkg.FailBy(err, "create replacement file with write_only fail.", nil)
	}

	f, err := os.OpenFile(file, os.O_RDWR|os.O_TRUNC, note.mode)
	defer f.Close()
	if err != nil {
		return errPkg.FailBy(err, "open file with read_write and trunc fail.", nil)
	}

	if _, err := io.Copy(nf, f); err != nil {
		nf.Close()
		os.Remove(note.newFile)
		return errPkg.FailBy(err, "exec copy: nf -> f fail.", errPkg.Fields{"f":file, "nf":note.newFile})
	}

    // https://stackoverflow.com/questions/44416645/truncate-a-file-in-golang
	f.Truncate(0)
	f.Seek(0, 0)
	return nil
}

func (note mockingExistentFileNote) back(file string) error {
	nf, err := os.OpenFile(note.newFile, os.O_RDONLY, note.mode)
	defer nf.Close()
	if err != nil {
		return errPkg.FailBy(err, "open replacement file with read_only fail.", nil)
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, note.mode)
	defer f.Close()
	if err != nil {
		return errPkg.FailBy(err, "re-create file with write_only fail.", nil)
	}

	if _, err := io.Copy(f, nf); err != nil {
		f.Close()
		os.Remove(file)
		return errPkg.FailBy(err, "exec copy: nf -> f fail.", errPkg.Fields{"f":file, "nf":note.newFile})
	}

	nf.Close()
	os.Remove(note.newFile)
	os.Chtimes(note.newFile, note.modTime, note.modTime)
	return nil
}

func (note mockingExistentFileNote) name() string {
	return "mock existent-file"
}

type mockingNonExistentFileNote struct {
	createdDirs []string
}

func mockNonExistentFile(file string) (*mockingNonExistentFileNote, error) {
	dirs := make([]string, 0)
	dir := filepath.Dir(file)
	for {
		if dir == "." || filepath.ToSlash(dir) == "/" {
			break
		}
		dirs = append(dirs, dir)
		dir = filepath.Dir(dir)
	}

	createdDirs := make([]string, 0)
	for i := len(dirs); i > 0; i-- {
		if _, err := os.Stat(dirs[i-1]); os.IsNotExist(err) {
			createdDirs = append(createdDirs, dirs[i-1])
		} else if err != nil {
			return nil, err
		}
	}

	return &mockingNonExistentFileNote{createdDirs: createdDirs}, nil
}

func (note mockingNonExistentFileNote) forward(file string) error {
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return errPkg.FailBy(err, "create file's dir fail.", nil)
	}

	f, err := os.OpenFile(file, os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		return errPkg.FailBy(err, "create file fail.", nil)
	}

	return nil
}

func (note mockingNonExistentFileNote) back(file string) error {
	for i:=len(note.createdDirs); i>0; i-- {
		if err := os.Remove(note.createdDirs[i-1]); err != nil {
			return errPkg.FailBy(err, "remove created dir fail.", nil)
		}
	}
	os.Remove(file)
	return nil
}

func (note mockingNonExistentFileNote) name() string {
	return "mock non-existent-file"
}

// https://stackoverflow.com/questions/14249217/how-do-i-know-im-running-within-go-test
func IsInTesting() bool {
	return flag.Lookup("test.v") != nil
}

func IgnoreCreatedAt(err error) error {
	known, ok := err.(*errPkg.Err)
	if !ok {
		return err
	}

	return &errPkg.Err{
		CreatedAt:time.Unix(0,0),
		Msg: known.Msg,
		Fields: known.Fields,
		Cause: known.Cause,
		StackInfo: known.StackInfo,
	}
}
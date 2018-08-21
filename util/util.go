package util

import (
	"net"
	"os"
	"path/filepath"
	"qing/go-helper/error"
	"fmt"
)

// https://stackoverflow.com/questions/18537257/how-to-get-the-directory-of-the-currently-running-file
func GetRuntimeDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}

// https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
//func FileExists(file string) (bool, error) {
//	_, err := os.Stat(file)
//	if err == nil {
//		return true, nil
//	}
//	if os.IsNotExist(err) {
//		return false, nil
//	}
//	return false, err
//}

// https://stackoverflow.com/questions/35558787/create-an-empty-text-file/35558965
// func CreateFileIfNotExist(file string) (*os.File, error)
// ==> os.OpenFile(file, os.CREATE|os.TRUNC, 0666)
// ==> os.Create(file)

// https://siongui.github.io/2017/03/28/go-create-directory-if-not-exist/
// https://stackoverflow.com/questions/37932551/mkdir-if-not-exists-using-golang
func CreateDirIfNotExist(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return err
}

// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func GetIntranetIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", errPkg.FailBy(err, "query system net interfaces fail", nil)
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", errPkg.FailBy(err, "query unicast interface addresses fail",
				errPkg.Fields{"interface":fmt.Sprintf("%+v", i)})
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.IsGlobalUnicast() {
				return ip.String(), nil
			}
		}
	}
	return "", errPkg.Fail("query global unicate ip fail", errPkg.Fields{"interfaces": ifaces})
}

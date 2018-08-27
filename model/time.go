package model

import (
	"time"
	"fmt"
)

var CST *time.Location

func init() {
	var err error
	CST, err = time.LoadLocation("Asia/Chongqing")
	if err != nil {
		panic(fmt.Sprintf("init <CST: china standard time> time.location fail: %s", err.Error()))
	}
}

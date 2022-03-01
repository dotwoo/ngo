package g

import (
	"fmt"
	"os"

	"github.com/NetEase-Media/ngo/util"
)

// Version 版本信息
// 编译时需要添加,如:
// -X "github.com/NetEase-Media/ngo/g.Version=v0.0.1" -X "github.com/NetEase-Media/ngo/g.BuildTime=2022-02-14 02:33:19 +0000" -X "github.com/NetEase-Media/ngo/g.Commit=44d77c8" -X "github.com/NetEase-Media/ngo/g.ProgName=ngo"
var (
	Version   = "unknown version"
	BuildTime = "unknown time"
	ProgName  = "unknown"
	Commit    = "unknown"
)

// PrintVersion 版本打印
func PrintVersion() {
	if !util.Containt(os.Args[1:], "-version") {
		return
	}
	fmt.Println("program:", ProgName)
	fmt.Println("version:", Version)
	fmt.Println("commit:", Commit)
	fmt.Println("buildtime:", BuildTime)
	os.Exit(0)
}

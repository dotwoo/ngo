// Description:
// Author: dotwoo@gmail.com
// Github: https://github.com/dotwoo
// Date: 2022-03-01 15:35:16
// FilePath: /ngo/cmd/main.go
//
package main

import (
	"context"

	"github.com/NetEase-Media/ngo/g"
	"github.com/NetEase-Media/ngo/internal/server"
	"github.com/NetEase-Media/ngo/pkg/adapter/log"
	"github.com/NetEase-Media/ngo/pkg/adapter/protocol"
	"github.com/gin-gonic/gin"
)

// go run . -c ./app.yaml
func main() {
	g.PrintVersion()

	s := server.Init()
	s.PreStart = func() error {
		log.Info("do pre-start...")
		return nil
	}

	s.PreStop = func(ctx context.Context) error {
		log.Info("do pre-stop...")
		return nil
	}

	s.AddRoute(server.GET, "/hello", func(ctx *gin.Context) {
		ctx.JSON(protocol.JsonBody("hello"))
	})
	s.Start()
}

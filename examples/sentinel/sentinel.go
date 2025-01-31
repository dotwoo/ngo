// Copyright Ngo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/NetEase-Media/ngo/internal/server"
	"github.com/NetEase-Media/ngo/pkg/adapter/protocol"
	"github.com/NetEase-Media/ngo/pkg/adapter/sentinel"
	"github.com/gin-gonic/gin"
)

// go run . -c ./app.yaml
func main() {
	s := server.Init()
	s.AddRoute(server.GET, "/hello", func(ctx *gin.Context) {

		e, b := sentinel.Entry("abc")
		if b != nil {
			ctx.String(429, "too many requests")
			return
		} else {
			e.Exit()
		}
		ctx.JSON(protocol.JsonBody("hello"))
		return
	})
	s.Start()
}

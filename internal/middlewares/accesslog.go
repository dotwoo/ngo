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

package middlewares

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/NetEase-Media/ngo/internal/middlewares/accesslog"

	"github.com/gin-gonic/gin"
)

type AccessLogMwOptions struct {
	Enabled         bool
	Pattern         string
	Path            string
	FileName        string
	FilePathPattern string // 定义文件路径名称格式
	NoFile          bool
	MaxCount        uint          // 默认3*24
	RotationTime    time.Duration // 默认1小时
	RotationSize    int64         // 单位MB，默认100
}

func NewDefaultAccessLogOptions() *AccessLogMwOptions {
	return &AccessLogMwOptions{
		Enabled:         true,
		Pattern:         accesslog.ApacheCombinedLogFormat,
		Path:            "",
		FileName:        "access",
		FilePathPattern: "",
		NoFile:          true,
		MaxCount:        72,
		RotationTime:    time.Hour,
		RotationSize:    100,
	}
}

func AccessLogMiddleware(opt *AccessLogMwOptions) gin.HandlerFunc {
	if opt == nil {
		opt = NewDefaultAccessLogOptions()
	}
	if opt.Enabled {
		if opt.NoFile {
			return accesslog.FormatWith(opt.Pattern, accesslog.WithOutput(os.Stdout))
		}

		writer, err := newRotateLog(opt)
		if err != nil {
			panic(err)
		}
		return accesslog.FormatWith(opt.Pattern, accesslog.WithOutput(writer))
	}
	return func(c *gin.Context) {
		c.Next()
	}
}

func newRotateLog(opt *AccessLogMwOptions) (io.Writer, error) {
	dir, err := filepath.Abs(opt.Path)
	if err != nil {
		return nil, err
	}
	//有需求要自定义filepattern
	var pathPattern string
	if len(opt.FilePathPattern) > 3 {
		pathPattern = opt.FilePathPattern
	} else {
		pathPattern = path.Join(dir, opt.FileName+".%Y-%m-%d-%H-%M.access.log")
	}
	linkName := path.Join(dir, opt.FileName+".access.log")
	return rotatelogs.New(
		pathPattern,
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.WithLinkName(linkName),
		rotatelogs.WithRotationTime(opt.RotationTime),
		rotatelogs.WithRotationCount(opt.MaxCount),
		rotatelogs.WithRotationSize(opt.RotationSize*1024*1024),
	)
}

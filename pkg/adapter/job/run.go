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

package job

import (
	"os"

	"github.com/NetEase-Media/ngo/pkg/adapter/config"
	"github.com/NetEase-Media/ngo/pkg/util"
)

// Run 运行函数并退出进程
func Run(f Callback) {
	var opt Options
	err := config.Unmarshal("job", &opt)
	util.CheckError(err)
	opt.check()

	hostname, err := os.Hostname()
	util.CheckError(err)

	j := &job{
		opt:      &opt,
		f:        f,
		hostname: hostname,
	}
	j.run()

	os.Exit(0)
}

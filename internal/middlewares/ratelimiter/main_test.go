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

package ratelimiter

import (
	"os"
	"testing"

	"github.com/NetEase-Media/ngo/pkg/adapter/sentinel"
	"github.com/alibaba/sentinel-golang/core/flow"
)

func TestMain(m *testing.M) {
	setupTest()
	ret := m.Run()
	tearDownTest()
	os.Exit(ret)
}

func setupTest() {
	sentinel.Init(&sentinel.Options{
		FlowRules: []*flow.Rule{
			{
				Resource:               "abc",
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				Threshold:              1,
				StatIntervalInMs:       1000,
			},
			{
				Resource:               "def",
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
				Threshold:              1,
				StatIntervalInMs:       1000,
			},
		},
	})
}

func tearDownTest() {

}

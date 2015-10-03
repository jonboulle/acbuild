// Copyright 2015 The appc Authors
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

package libacb

import (
	"github.com/appc/acbuild/libacb/util"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
)

func removeFromEnv(name string) func(*schema.ImageManifest) {
	return func(s *schema.ImageManifest) {
		if s.App == nil {
			return
		}
		for i := len(s.App.Environment) - 1; i >= 0; i-- {
			if s.App.Environment[i].Name == name {
				s.App.Environment = append(
					s.App.Environment[:i],
					s.App.Environment[i+1:]...)
			}
		}
	}
}

func AddEnv(acipath, name, value string) error {
	fn := func(s *schema.ImageManifest) {
		if s.App == nil {
			s.App = &types.App{}
		}
		s.App.Environment.Set(name, value)
	}
	return util.ModifyManifest(fn, acipath)
}

func RemoveEnv(acipath, name string) error {
	return util.ModifyManifest(removeFromEnv(name), acipath)
}

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

func removeLabelFromMan(name types.ACIdentifier) func(*schema.ImageManifest) {
	return func(s *schema.ImageManifest) {
		for i := len(s.Labels) - 1; i >= 0; i-- {
			if s.Labels[i].Name == name {
				s.Labels = append(
					s.Labels[:i],
					s.Labels[i+1:]...)
			}
		}
	}
}

func AddLabel(acipath, name, value string) error {
	acid, err := types.NewACIdentifier(name)
	if err != nil {
		return err
	}

	fn := func(s *schema.ImageManifest) {
		removeLabelFromMan(*acid)(s)
		s.Labels = append(s.Labels,
			types.Label{
				Name:  *acid,
				Value: value,
			})
	}
	return util.ModifyManifest(fn, acipath)
}

func RemoveLabel(acipath, name string) error {
	acid, err := types.NewACIdentifier(name)
	if err != nil {
		return err
	}

	return util.ModifyManifest(removeLabelFromMan(*acid), acipath)
}

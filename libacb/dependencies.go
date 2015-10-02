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

func removeDep(imageName types.ACIdentifier) func(*schema.ImageManifest) {
	return func(s *schema.ImageManifest) {
		for i := len(s.Dependencies) - 1; i >= 0; i-- {
			if s.Dependencies[i].ImageName == imageName {
				s.Dependencies = append(
					s.Dependencies[:i],
					s.Dependencies[i+1:]...)
			}
		}
	}
}

func AddDependency(acipath, imageName, imageId string, labels types.Labels, size uint) error {
	acid, err := types.NewACIdentifier(imageName)
	if err != nil {
		return err
	}

	var hash *types.Hash
	if imageId != "" {
		var err error
		hash, err = types.NewHash(imageId)
		if err != nil {
			return err
		}
	}

	fn := func(s *schema.ImageManifest) {
		removeDep(*acid)(s)
		s.Dependencies = append(s.Dependencies,
			types.Dependency{
				ImageName: *acid,
				ImageID:   hash,
				Labels:    labels,
				Size:      size,
			})
	}
	return util.ModifyManifest(fn, acipath)
}

func RemoveDependency(acipath, imageName string) error {
	acid, err := types.NewACIdentifier(imageName)
	if err != nil {
		return err
	}

	return util.ModifyManifest(removeDep(*acid), acipath)
}

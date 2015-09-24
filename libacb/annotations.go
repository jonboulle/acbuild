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

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
)

func removeAnnotation(name types.ACIdentifier) func(*schema.ImageManifest) {
	return func(s *schema.ImageManifest) {
		for i := len(s.Annotations) - 1; i >= 0; i-- {
			if s.Annotations[i].Name == name {
				s.Annotations = append(
					s.Annotations[:i],
					s.Annotations[i+1:]...)
			}
		}
	}
}

func AddAnnotation(acipath, name, value string) error {
	acid, err := types.NewACIdentifier(name)
	if err != nil {
		return err
	}

	fn := func(s *schema.ImageManifest) {
		removeAnnotation(*acid)(s)
		s.Annotations = append(s.Annotations,
			types.Annotation{
				Name:  *acid,
				Value: value,
			})
	}
	return util.ModifyManifest(fn, acipath)
}

func RemoveAnnotation(acipath, name string) error {
	acid, err := types.NewACIdentifier(name)
	if err != nil {
		return err
	}

	return util.ModifyManifest(removeAnnotation(*acid), acipath)
}

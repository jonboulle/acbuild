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
	"fmt"
	"os"
	"path"

	"github.com/appc/acbuild/libacb/util"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
)

var (
	placeholderexec = []string{"/acbuild", "placeholder", "exec", "statement"}
	placeholdername = "acbuild-unnamed"
)

func Begin(tmpaci string, start *string) error {
	ex, err := util.Exists(tmpaci)
	if err != nil {
		return err
	}
	if ex {
		return fmt.Errorf("build already in progress? path exists: %s", tmpaci)
	}

	err = os.MkdirAll(path.Join(tmpaci, "rootfs"), 0755)
	if err != nil {
		return err
	}

	if start != nil {
		ex, err := util.Exists(*start)
		if err != nil {
			return err
		}
		if !ex {
			return fmt.Errorf("start aci doesn't exist: %s", *start)
		}

		err = util.ExpandACI(*start, tmpaci)
		if err != nil {
			return err
		}

		return nil
	}

	acid, err := types.NewACIdentifier("acbuild-unnamed")
	if err != nil {
		return err
	}

	manifest := &schema.ImageManifest{
		ACKind:    schema.ImageManifestKind,
		ACVersion: schema.AppContainerVersion,
		Name:      *acid,
		App: &types.App{
			Exec:  placeholderexec,
			User:  "0",
			Group: "0",
		},
	}

	manblob, err := manifest.MarshalJSON()
	if err != nil {
		err = util.RmIfPossible(tmpaci, err)
		return err
	}

	manfile, err := os.Create(path.Join(tmpaci, "manifest"))
	if err != nil {
		err = util.RmIfPossible(tmpaci, err)
		return err
	}

	_, err = manfile.Write(manblob)
	if err != nil {
		err = util.RmIfPossible(tmpaci, err)
		return err
	}

	err = manfile.Close()
	if err != nil {
		err = util.RmIfPossible(tmpaci, err)
		return err
	}

	return nil
}

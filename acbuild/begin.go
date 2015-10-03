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

package main

import (
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/libacb"
)

var (
	cmdBegin = &cobra.Command{
		Use:     "begin [START_ACI_PATH]",
		Short:   "Start a new build",
		Long:    "Begins a new build. By default operations will be performed on top of an empty image, but a start image can be provided",
		Example: "acbuild begin",
		Run:     runWrapper(runBegin),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdBegin)
}

func runBegin(cmd *cobra.Command, args []string) (exit int) {
	if len(args) > 1 {
		stderr("begin: too many arguments")
		return 1
	}

	if debug {
		if len(args) == 0 {
			stderr("Beginning build with an empty ACI")
		} else {
			stderr("Beginning build with %s", args[0])
		}
	}

	var err error
	if len(args) == 0 {
		err = libacb.Begin(tmpacipath(), "")
	} else {
		err = libacb.Begin(tmpacipath(), args[0])
	}

	if err != nil {
		stderr("begin: %v", err)
		return 1
	}

	return 0
}

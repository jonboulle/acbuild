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
	insecure = false
	cmdExec  = &cobra.Command{
		Use:     "exec CMD [ARGS]",
		Short:   "Run a command in an ACI",
		Long:    "Run a given command in an ACI, and save the resulting container as a new ACI",
		Example: "acbuild exec yum install nginx",
		Run:     runWrapper(runExec),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdExec)

	cmdExec.Flags().BoolVar(&insecure, "insecure", false, "Allows fetching dependencies over http")
}

func runExec(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}

	if debug {
		stderr("Execing: %v", args)
	}

	err := libacb.Exec(tmpacipath(), depstorepath(), targetpath(),
		scratchpath(), workpath(), args, insecure, debug)

	if err != nil {
		stderr("exec: %v", err)
		return 1
	}

	return 0
}

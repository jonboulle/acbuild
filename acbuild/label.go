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
	cmdLabel = &cobra.Command{
		Use:   "label [command]",
		Short: "Manage labels",
	}
	cmdAddLabel = &cobra.Command{
		Use:     "add NAME VALUE",
		Short:   "Add a label",
		Long:    "Updates the ACI to contain a label with the given name and value. If the label already exists, its value will be changed.",
		Example: "acbuild label add arch amd64",
		Run:     runWrapper(runAddLabel),
	}
	cmdRmLabel = &cobra.Command{
		Use:     "remove NAME",
		Aliases: []string{"rm"},
		Short:   "Remove a label",
		Long:    "Updates the labels in the ACI's manifest to not include the label for the given name",
		Example: "acbuild label remove arch",
		Run:     runWrapper(runRemoveLabel),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdLabel)
	cmdLabel.AddCommand(cmdAddLabel)
	cmdLabel.AddCommand(cmdRmLabel)
}

func runAddLabel(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}
	if len(args) != 2 {
		stderr("add-label: incorrect number of arguments")
		return 1
	}

	if debug {
		stderr("Adding label \"%s\"=\"%s\"", args[0], args[1])
	}

	err := libacb.AddLabel(tmpaci(), args[0], args[1])

	if err != nil {
		stderr("add-label: %v", err)
		return 1
	}

	return 0
}

func runRemoveLabel(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}
	if len(args) != 1 {
		stderr("rm-label: too many arguments")
		return 1
	}

	if debug {
		stderr("Removing label \"%s\"", args[0])
	}

	err := libacb.RemoveLabel(tmpaci(), args[0])

	if err != nil {
		stderr("rm-label: %v", err)
		return 1
	}

	return 0
}

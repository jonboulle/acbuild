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
	"github.com/spf13/cobra"

	"github.com/appc/acbuild/libacb"
)

var (
	readOnly bool
	cmdMount = &cobra.Command{
		Use:   "mount [command]",
		Short: "Manage mount points",
	}
	cmdAddMount = &cobra.Command{
		Use:     "add NAME PATH",
		Short:   "Add a mount point",
		Long:    "Updates the ACI to contain a mount point with the given name and path. If the mount point already exists, its path will be changed.",
		Example: "acbuild mount add htmlfiles /usr/share/nginx/html --read-only",
		Run:     runWrapper(runAddMount),
	}
	cmdRmMount = &cobra.Command{
		Use:     "remove NAME",
		Aliases: []string{"rm"},
		Short:   "Remove a mount point",
		Long:    "Removes the mount point with the given name from the ACI's manifest",
		Example: "acbuild mount remove htmlfiles",
		Run:     runWrapper(runRmMount),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdMount)
	cmdMount.AddCommand(cmdAddMount)
	cmdMount.AddCommand(cmdRmMount)

	cmdAddMount.Flags().BoolVar(&readOnly, "read-only", false, "Set the mount point to be read only")
}

func runAddMount(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}
	if len(args) != 2 {
		stderr("add-mount: incorrect number of arguments")
		return 1
	}

	if debug {
		if readOnly {
			stderr("Adding read only mount point \"%s\"=\"%s\"", args[0], args[1])
		} else {
			stderr("Adding mount point \"%s\"=\"%s\"", args[0], args[1])
		}
	}

	err := libacb.AddMount(tmpaci(), args[0], args[1], readOnly)

	if err != nil {
		stderr("add-mount: %v", err)
		return 1
	}

	return 0
}

func runRmMount(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}
	if len(args) != 1 {
		stderr("rm-mount: too many arguments")
		return 1
	}

	if debug {
		stderr("Removing mount point \"%s\"", args[0])
	}

	err := libacb.RemoveMount(tmpaci(), args[0])

	if err != nil {
		stderr("rm-mount: %v", err)
		return 1
	}

	return 0
}

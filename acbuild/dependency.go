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
	"fmt"
	"strings"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/libacb"
)

var (
	imageId string
	labels  labellist
	size    uint
	cmdDep  = &cobra.Command{
		Use:     "dependency [command]",
		Aliases: []string{"dep"},
		Short:   "Manage dependencies",
	}
	cmdAddDep = &cobra.Command{
		Use:     "add IMAGE_NAME",
		Short:   "Add a dependency",
		Long:    "Updates the ACI to contain a dependency with the given name. If the dependency already exists, its values will be changed.",
		Example: "acbuild dependency add example.com/reduce-worker-base --label os=linux --label env=canary --size 22017258",
		Run:     runWrapper(runAddDep),
	}
	cmdRmDep = &cobra.Command{
		Use:     "remove IMAGE_NAME",
		Aliases: []string{"rm"},
		Short:   "Remove a dependency",
		Long:    "Removes the dependency with the given name from the ACI's manifest",
		Example: "acbuild dependency remove example.com/reduce-worker-base",
		Run:     runWrapper(runRmDep),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdDep)
	cmdDep.AddCommand(cmdAddDep)
	cmdDep.AddCommand(cmdRmDep)

	cmdAddDep.Flags().StringVar(&imageId, "image-id", "", "Content hash of the dependency")
	cmdAddDep.Flags().Var(&labels, "label", "Labels used for dependency matching")
	cmdAddDep.Flags().UintVar(&size, "size", 0, "The size of the image referenced dependency, in bytes")
}

func runAddDep(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}
	if len(args) != 1 {
		stderr("dependency add: too many arguments")
		return 1
	}

	if debug {
		stderr("Adding dependency %q=%q", args[0], args[1])
	}

	appcLabels := make(types.Labels, len(labels))
	for i, label := range labels {
		appcLabels[i] = types.Label{
			Name:  types.ACIdentifier(label.Name),
			Value: label.Value,
		}
	}

	err := libacb.AddDependency(tmpacipath(), args[0], imageId, appcLabels, size)

	if err != nil {
		stderr("dependency add: %v", err)
		return 1
	}

	return 0
}

func runRmDep(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}
	if len(args) != 1 {
		stderr("dependency remove: too many arguments")
		return 1
	}

	if debug {
		stderr("Removing dependency %q", args[0])
	}

	err := libacb.RemoveDependency(tmpacipath(), args[0])

	if err != nil {
		stderr("dependency-remove: %v", err)
		return 1
	}

	return 0
}

type labellist []types.Label

func (ls *labellist) String() string {
	strLabels := make([]string, len(*ls))
	for i, label := range *ls {
		strLabels[i] = fmt.Sprintf("%s=%s", label.Name, label.Value)
	}
	return strings.Join(strLabels, " ")
}

func (ls *labellist) Set(input string) error {
	parts := strings.SplitN(input, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("no '=' character in %q", input)
	}
	*ls = append(*ls, types.Label{
		Name:  types.ACIdentifier(parts[0]),
		Value: parts[1],
	})
	return nil
}

func (ls *labellist) Type() string {
	return "Labels"
}

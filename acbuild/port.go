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
	"strconv"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/spf13/cobra"

	"github.com/appc/acbuild/libacb"
)

var (
	count           uint
	socketActivated bool
	cmdPort         = &cobra.Command{
		Use:   "port [command]",
		Short: "Manage ports",
	}
	cmdAddPort = &cobra.Command{
		Use:     "add NAME PROTOCOL PORT",
		Short:   "Add a port",
		Long:    "Updates the ACI to contain a port with the given name, protocol, and port. If the port already exists, its values will be changed.",
		Example: "acbuild port add https 443 tcp --socket-activated",
		Run:     runWrapper(runAddPort),
	}
	cmdRmPort = &cobra.Command{
		Use:     "remove NAME",
		Aliases: []string{"rm"},
		Short:   "Remove a port",
		Long:    "Updates the ports in the ACI's manifest to include a port with the given name and value",
		Example: "acbuild port remove https",
		Run:     runWrapper(runRmPort),
	}
)

func init() {
	cmdAcbuild.AddCommand(cmdPort)
	cmdPort.AddCommand(cmdAddPort)
	cmdPort.AddCommand(cmdRmPort)

	cmdAddPort.Flags().UintVar(&count, "count", 1, "Specifies a range of ports, going from PORT to PORT + count - 1")
	cmdAddPort.Flags().BoolVar(&socketActivated, "socket-activated", false, "Set the app to be socket activated on this/these port/ports")
}

func runAddPort(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}
	if len(args) != 3 {
		stderr("add-port: incorrect number of arguments")
		return 1
	}
	port, err := strconv.ParseUint(args[2], 10, 16)
	if err != nil {
		stderr("add-port: port must be a positive number between 0 and 65535")
		return 1
	}

	if debug {
		stderr("Adding port \"%s\"=\"%s\"", args[0], args[1])
	}

	err = libacb.AddPort(tmpaci(), args[0], args[1], uint(port), count, socketActivated)

	if err != nil {
		stderr("add-port: %v", err)
		return 1
	}

	return 0
}

func runRmPort(cmd *cobra.Command, args []string) (exit int) {
	if len(args) == 0 {
		cmd.Usage()
		return 1
	}
	if len(args) != 1 {
		stderr("rm-port: too many arguments")
		return 1
	}

	if debug {
		stderr("Removing port \"%s\"", args[0])
	}

	err := libacb.RemovePort(tmpaci(), args[0])

	if err != nil {
		stderr("rm-port: %v", err)
		return 1
	}

	return 0
}

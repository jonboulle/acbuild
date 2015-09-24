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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/appc/acbuild/libacb/util"
)

var pathlist = []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin"}

func Exec(acipath, depstore, targetpath, scratchpath, workpath string, cmd []string, insecure bool) error {
	err := util.RmAndMkdir(targetpath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(targetpath)
	err = util.RmAndMkdir(workpath)
	if err != nil {
		return err
	}
	defer os.RemoveAll(workpath)
	err = util.MkdirIfMissing(scratchpath)
	if err != nil {
		return err
	}
	err = util.MkdirIfMissing(depstore)
	if err != nil {
		return err
	}

	man, err := util.GetManifest(acipath)
	if err != nil {
		return err
	}

	if len(man.Dependencies) != 0 {
		err := util.Exec("modprobe", "overlay")
		if err != nil {
			return err
		}
		if !supportsOverlay() {
			return fmt.Errorf("overlayfs support required for using exec with dependencies")
		}
	}

	deps, err := util.RenderACI(acipath, scratchpath, depstore, insecure)
	if err != nil {
		return err
	}

	var nspawnpath string
	if deps == nil {
		nspawnpath = path.Join(acipath, "rootfs")
	} else {
		//TODO: escape colons
		for i, dep := range deps {
			deps[i] = path.Join(dep, "/rootfs")
		}
		options := "-olowerdir=" + strings.Join(deps, ":") + ",upperdir=" + path.Join(acipath, "/rootfs") + ",workdir=" + workpath
		err := util.Exec("mount", "-t", "overlay", "overlay", options, targetpath)
		if err != nil {
			return err
		}

		umount := exec.Command("umount", targetpath)
		umount.Stdout = os.Stdout
		umount.Stderr = os.Stderr
		defer umount.Run()

		nspawnpath = targetpath
	}
	nspawncmd := []string{"systemd-nspawn", "-q", "-D", nspawnpath}

	if man.App != nil {
		for _, evar := range man.App.Environment {
			nspawncmd = append(nspawncmd, "--setenv", evar.Name+"="+evar.Value)
		}
	}

	if len(cmd) == 0 {
		return fmt.Errorf("command to run not set")
	}
	abscmd, err := findCmdInPath(pathlist, cmd[0], nspawnpath)
	if err != nil {
		return err
	}
	nspawncmd = append(nspawncmd, abscmd)
	nspawncmd = append(nspawncmd, cmd[1:]...)
	//fmt.Printf("%v\n", nspawncmd)

	err = util.Exec(nspawncmd[0], nspawncmd[1:]...)
	if err != nil {
		return err
	}

	return nil
}

// stolen from github.com/coreos/rkt/common/common.go
// supportsOverlay returns whether the system supports overlay filesystem
func supportsOverlay() bool {
	f, err := os.Open("/proc/filesystems")
	if err != nil {
		fmt.Println("error opening /proc/filesystems")
		return false
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == "nodev\toverlay" {
			return true
		}
	}
	return false
}

func findCmdInPath(pathlist []string, cmd, prefix string) (string, error) {
	if len(cmd) != 0 && cmd[0] == '/' {
		return cmd, nil
	}

	for _, p := range pathlist {
		ex, err := util.Exists(path.Join(prefix, p, cmd))
		if err != nil {
			return "", err
		}
		if ex {
			return path.Join(p, cmd), nil
		}
	}
	return "", fmt.Errorf("%s not found in any of: %v", cmd, pathlist)
}

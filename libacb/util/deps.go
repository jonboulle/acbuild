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

package util

import (
	"crypto/sha512"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/discovery"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/ioprogress"
)

var (
	defaultArch = "amd64"
	defaultOS   = "linux"
)

// RenderACI will download any missing dependencies into the depstore,
// expand any unexpanded dependencies into the scratchpath, and return
// an ordered list of all dependencies for overlaying.
func RenderACI(acipath, scratchpath, depstore string, insecure bool) ([]string, error) {
	man, err := GetManifest(acipath)
	if err != nil {
		return nil, err
	}

	if len(man.Dependencies) == 0 {
		return nil, nil
	}

	remainingPaths := []string{acipath}
	index := 0

	for len(remainingPaths) > index {
		depPaths, err := renderACI(remainingPaths[index], scratchpath, depstore, insecure)
		if err != nil {
			return nil, err
		}

		tmpList := append(remainingPaths[:index+1], depPaths...)
		remainingPaths = append(tmpList, remainingPaths[index+1:]...)
		index++
	}

	return remainingPaths[1:], nil
}

func renderACI(acipath, scratchpath, depstore string, insecure bool) ([]string, error) {
	man, err := GetManifest(acipath)
	if err != nil {
		return nil, err
	}

	err = FetchDeps(man.Dependencies, depstore, insecure)
	if err != nil {
		return nil, err
	}

	var expandedDeps []string
	for _, dep := range man.Dependencies {
		depname := string(dep.ImageName)
		expandedpath := GenDepPath(scratchpath, depname)

		ex, err := Exists(expandedpath)
		if err != nil {
			return nil, err
		}
		if ex {
			expandedDeps = append(expandedDeps, expandedpath)
			continue
		}

		err = os.MkdirAll(expandedpath, 0755)
		if err != nil {
			return nil, err
		}

		err = ExpandACI(GenDepPath(depstore, depname), expandedpath)
		if err != nil {
			os.RemoveAll(expandedpath)
			return nil, err
		}

		expandedDeps = append(expandedDeps, expandedpath)
	}

	reversedDeps := make([]string, len(expandedDeps))
	for i, dep := range expandedDeps {
		reversedDeps[len(expandedDeps)-i-1] = dep
	}

	return reversedDeps, nil
}

func ExpandACI(acipath, targetpath string) error {
	//TODO: there should be a way to avoid exec'ing another binary here
	cmd := exec.Command("tar", "xf", acipath, "-C", targetpath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

//TODO: path should include labels probably
func GenDepPath(depstore, imageName string) string {
	return path.Join(depstore, strings.Replace(imageName, "/", "-", -1)+".aci")
}

func FetchDeps(deps types.Dependencies, depstore string, insecure bool) error {
	for _, dep := range deps {
		err := MkdirIfMissing(depstore)
		if err != nil {
			return err
		}

		deppath := GenDepPath(depstore, string(dep.ImageName))

		ex, err := Exists(deppath)
		if err != nil {
			return err
		}
		if ex {
			if dep.ImageID != nil {
				matches, err := VerifyACIHash(deppath, *dep.ImageID)
				if err != nil {
					return err
				}

				if !matches {
					return fmt.Errorf("%s doesn't match hash", dep.ImageName)
				}
			}
			if dep.Size != 0 {
				depfile, err := os.Open(deppath)
				if err != nil {
					return err
				}

				depstat, err := depfile.Stat()
				if err != nil {
					err2 := depfile.Close()
					if err2 != nil {
						return fmt.Errorf("error stating: %v, closing: %v", err, err2)
					}
					return err
				}

				err = depfile.Close()
				if err != nil {
					return err
				}

				if dep.Size != uint(depstat.Size()) {
					return fmt.Errorf("%s doesn't match size, expected: %d, actual %d",
						dep.ImageName, dep.Size, depstat.Size())
				}
			}
			//TODO: check if labels match
			continue
		}

		app, err := discovery.NewAppFromString(string(dep.ImageName))
		if err != nil {
			return err
		}

		if _, ok := app.Labels["arch"]; !ok {
			app.Labels["arch"] = defaultArch
		}
		if _, ok := app.Labels["os"]; !ok {
			app.Labels["os"] = defaultOS
		}

		eps, _, err := discovery.DiscoverEndpoints(*app, insecure)
		//for _, a := range attempts {
		//	fmt.Printf("meta tag not found on %s: %v\n", a.Prefix, a.Error)
		//}
		if err != nil {
			return err
		}

		if len(eps.ACIEndpoints) == 0 {
			return fmt.Errorf("no endpoints discovered to download %s", dep.ImageName)
		}

		endpoint := eps.ACIEndpoints[0]

		err = Download(endpoint.ACI, deppath, string(dep.ImageName), insecure)
		if err != nil {
			return err
		}

		err = Download(endpoint.ASC, deppath+".asc", "Signature", insecure)
		if err != nil {
			return err
		}

		//TODO: verify the downloaded image
	}

	return nil
}

func VerifyACIHash(path string, dephash types.Hash) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}

	h := sha512.New()

	_, err = io.Copy(h, file)
	if err != nil {
		return false, err
	}

	err = file.Close()
	if err != nil {
		return false, err
	}

	filehash := h.Sum(nil)

	return string(filehash) == dephash.Val, nil
}

func Download(url, path, label string, insecure bool) error {
	//TODO: auth
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	transport := http.DefaultTransport
	if insecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client := &http.Client{Transport: transport}
	//f.setHTTPHeaders(req, etag)

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}
		//f.setHTTPHeaders(req, etag)
		return nil
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		break
	default:
		return fmt.Errorf("bad HTTP status code: %d", res.StatusCode)
	}

	prefix := "Downloading " + label
	fmtBytesSize := 18
	barSize := int64(80 - len(prefix) - fmtBytesSize)
	bar := ioprogress.DrawTextFormatBarForW(barSize, os.Stderr)
	fmtfunc := func(progress, total int64) string {
		// Content-Length is set to -1 when unknown.
		if total == -1 {
			return fmt.Sprintf(
				"%s: %v of an unknown total size",
				prefix,
				ioprogress.ByteUnitStr(progress),
			)
		}
		return fmt.Sprintf(
			"%s: %s %s",
			prefix,
			bar(progress, total),
			ioprogress.DrawTextFormatBytes(progress, total),
		)
	}

	reader := &ioprogress.Reader{
		Reader:       res.Body,
		Size:         res.ContentLength,
		DrawFunc:     ioprogress.DrawTerminalf(os.Stderr, fmtfunc),
		DrawInterval: time.Second,
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, reader)
	if err != nil {
		return fmt.Errorf("error copying %s: %v", label, err)
	}

	err = out.Sync()
	if err != nil {
		return fmt.Errorf("error writing %s: %v", label, err)
	}

	err = out.Close()
	if err != nil {
		return err
	}

	return nil
}

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

package registry

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha512"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/aci"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/discovery"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/pkg/acirenderer"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/appc/acbuild/Godeps/_workspace/src/xi2.org/x/xz"

	"github.com/appc/acbuild/libacb/util"
)

var (
	defaultArch = "amd64"
	defaultOS   = "linux"
)

func (r Registry) tmppath() string {
	return path.Join(r.Depstore, "tmp.aci")
}

func (r Registry) tmpuncompressedpath() string {
	return path.Join(r.Depstore, "tmp.uncompressed.aci")
}

// FetchAndRender will fetch the given image and all of its dependencies if
// they have not been fetched yet, and will then render them on to the
// filesystem if they have not been rendered yet.
func (r Registry) FetchAndRender(imagename types.ACIdentifier, labels types.Labels, size uint) error {
	_, err := r.GetACI(imagename, labels)
	if err == ErrNotFound {
		err := r.fetchACIWithSize(imagename, labels, size)
		if err != nil {
			return err
		}
	}

	filesToRender, err := acirenderer.GetRenderedACI(imagename,
		labels, r)
	if err != nil {
		return err
	}

	for _, fs := range filesToRender {
		ex, err := util.Exists(path.Join(r.Scratchpath, fs.Key, "rendered"))
		if err != nil {
			return err
		}
		if ex {
			//This ACI has already been rendered
			continue
		}

		err = util.UnTar(path.Join(r.Depstore, fs.Key),
			path.Join(r.Scratchpath, fs.Key), fs.FileMap, string(imagename),
			r.Debug)
		if err != nil {
			return err
		}

		rfile, err := os.Create(path.Join(r.Scratchpath, fs.Key, "rendered"))
		if err != nil {
			return err
		}
		rfile.Close()
	}
	return nil
}

func (r Registry) fetchACIWithSize(imagename types.ACIdentifier, labels types.Labels, size uint) error {
	endpoint, err := r.discoverEndpoint(imagename, labels)
	if err != nil {
		return err
	}

	err = r.download(endpoint.ACI, r.tmppath(), string(imagename))
	if err != nil {
		return err
	}

	//TODO: download .asc, verify the .aci with it

	if size != 0 {
		finfo, err := os.Stat(r.tmppath())
		if err != nil {
			return err
		}
		if finfo.Size() != int64(size) {
			return fmt.Errorf(
				"dependency %s has incorrect size: expected=%d, actual=%d",
				size, finfo.Size())
		}
	}

	err = r.uncompress(string(imagename))
	if err != nil {
		return err
	}

	err = os.Remove(r.tmppath())
	if err != nil {
		return err
	}

	id, err := GenImageID(r.tmpuncompressedpath())
	if err != nil {
		return err
	}

	err = os.Rename(r.tmpuncompressedpath(), path.Join(r.Depstore, id))
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(r.Scratchpath, id, "rootfs"), 0755)
	if err != nil {
		return err
	}

	err = getManifestFromTar(path.Join(r.Depstore, id),
		path.Join(r.Scratchpath, id, "manifest"))
	if err != nil {
		return err
	}

	man, err := r.GetImageManifest(id)
	if err != nil {
		return err
	}

	if man.Name != imagename {
		return fmt.Errorf(
			"downloaded ACI name %q does not match expected image name %q",
			man.Name, imagename)
	}

	for _, dep := range man.Dependencies {
		err := r.fetchACIWithSize(dep.ImageName, dep.Labels, dep.Size)
		if err != nil {
			return err
		}
		if dep.ImageID != nil {
			id, err := r.GetACI(dep.ImageName, dep.Labels)
			if err != nil {
				return err
			}
			if id != dep.ImageID.String() {
				return fmt.Errorf("dependency %s doesn't match hash",
					dep.ImageName)
			}
		}
	}
	return nil
}

// Need to uncompress the file to be able to generate the Image ID
func (r Registry) uncompress(name string) error {
	acifile, err := os.Open(r.tmppath())
	if err != nil {
		return err
	}
	defer acifile.Close()

	typ, err := aci.DetectFileType(acifile)
	if err != nil {
		return err
	}

	// In case DetectFileType changed the cursor
	_, err = acifile.Seek(0, 0)
	if err != nil {
		return err
	}

	var in io.Reader

	if r.Debug {
		finfo, err := acifile.Stat()
		if err != nil {
			return err
		}
		in = util.NewIoprogress(fmt.Sprintf("Uncompressing %s", name),
			finfo.Size(), acifile)
	} else {
		in = acifile
	}

	switch typ {
	case aci.TypeGzip:
		in, err = gzip.NewReader(in)
		if err != nil {
			return err
		}
	case aci.TypeBzip2:
		in = bzip2.NewReader(in)
	case aci.TypeXz:
		in, err = xz.NewReader(in, 0)
		if err != nil {
			return err
		}
	case aci.TypeTar:
		break
	case aci.TypeText:
		return fmt.Errorf("downloaded ACI is text, not a tarball")
	case aci.TypeUnknown:
		return fmt.Errorf("downloaded ACI is of an unknown type")
	}

	out, err := os.OpenFile(r.tmpuncompressedpath(),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("error copying: %v", err)
	}

	err = out.Sync()
	if err != nil {
		return fmt.Errorf("error writing: %v", err)
	}

	return nil
}

func getManifestFromTar(tarpath, dst string) error {
	tarfile, err := os.Open(tarpath)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tr := tar.NewReader(tarfile)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// End of tar reached
			break
		}
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			if hdr.Name == "manifest" {
				f, err := os.OpenFile(dst,
					os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				defer f.Close()
				_, err = io.Copy(f, tr)
				if err != nil {
					return err
				}
			}
		default:
			continue
		}
	}
	return nil
}

func GenImageID(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha512.New()

	_, err = io.Copy(h, file)
	if err != nil {
		return "", err
	}

	s := h.Sum(nil)

	return fmt.Sprintf("sha512-%x", s), nil
}

func (r Registry) discoverEndpoint(imageName types.ACIdentifier, labels types.Labels) (*discovery.ACIEndpoint, error) {
	labelmap := make(map[types.ACIdentifier]string)
	for _, label := range labels {
		labelmap[label.Name] = label.Value
	}

	app, err := discovery.NewApp(string(imageName), labelmap)
	if err != nil {
		return nil, err
	}
	if _, ok := app.Labels["arch"]; !ok {
		app.Labels["arch"] = defaultArch
	}
	if _, ok := app.Labels["os"]; !ok {
		app.Labels["os"] = defaultOS
	}

	eps, attempts, err := discovery.DiscoverEndpoints(*app, r.Insecure)
	if err != nil {
		return nil, err
	}
	if r.Debug {
		for _, a := range attempts {
			fmt.Fprintf(os.Stderr, "meta tag not found on %s: %v\n",
				a.Prefix, a.Error)
		}
	}
	if len(eps.ACIEndpoints) == 0 {
		return nil, fmt.Errorf("no endpoints discovered to download %s",
			imageName)
	}

	return &eps.ACIEndpoints[0], nil
}

func (r Registry) download(url, path, label string) error {
	//TODO: auth
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	transport := http.DefaultTransport
	if r.Insecure {
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

	out, err := os.Create(path)
	if err != nil {
		return err
	}

	reader := util.NewIoprogress("Downloading "+label, res.ContentLength, res.Body)

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

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
	//"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/ioprogress"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/pkg/tar"
	"github.com/appc/acbuild/Godeps/_workspace/src/github.com/coreos/rkt/pkg/uid"
)

// MkdirIfMissing makes a directory at path if one does not exist.
func MkdirIfMissing(path string) error {
	ex, err := Exists(path)
	if err != nil {
		return err
	}
	if !ex {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// RmAndMkdir will remove anything at path if it exists, and then create a
// directory at path.
func RmAndMkdir(path string) error {
	ex, err := Exists(path)
	if err != nil {
		return err
	}
	if ex {
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

// Exists will return whether or not anything exists at path
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// NewIoprogress will wrap the provided io.Reader in an IOProgress
func NewIoprogress(label string, size int64, rdr io.Reader) io.Reader {
	fmtBytesSize := 18

	if len(label)+fmtBytesSize > 80 {
		return rdr
	}

	barSize := int64(80 - len(label) - fmtBytesSize)
	bar := ioprogress.DrawTextFormatBarForW(barSize, os.Stderr)
	fmtfunc := func(progress, total int64) string {
		// Content-Length is set to -1 when unknown.
		if total == -1 {
			return fmt.Sprintf(
				"%s: %v of an unknown total size",
				label,
				ioprogress.ByteUnitStr(progress),
			)
		}
		return fmt.Sprintf(
			"%s: %s %s",
			label,
			bar(progress, total),
			ioprogress.DrawTextFormatBytes(progress, total),
		)
	}

	return &ioprogress.Reader{
		Reader:       rdr,
		Size:         size,
		DrawFunc:     ioprogress.DrawTerminalf(os.Stderr, fmtfunc),
		DrawInterval: time.Second,
	}
}

// UnTar will extract the contents at the tar file at tarpath to the directory
// at dst. If fileMap is set, only files in it will be extracted.
func UnTar(tarpath, dst string, fileMap map[string]struct{}, name string, debug bool) error {
	dst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	tarfile, err := os.Open(tarpath)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	var in io.Reader
	in = tarfile
	if debug {
		finfo, err := tarfile.Stat()
		if err != nil {
			return err
		}
		in = NewIoprogress(fmt.Sprintf("Expanding %s", name), finfo.Size(),
			tarfile)
	}

	if fileMap == nil {
		fmt.Printf("fileMap is nil\n")
	} else {
		fmt.Printf("There are %d entries in the fileMap\n", len(fileMap))
	}

	return tar.ExtractTar(in, dst, true, uid.NewBlankUidRange(), fileMap)
}

//// UnTar will extract the contents at the tar file at tarpath to the directory
//// at dst. If fileMap is set, only files in it will be extracted.
//func UnTar(tarpath, dst string, fileMap map[string]struct{}) error {
//	tarfile, err := os.Open(tarpath)
//	if err != nil {
//		return err
//	}
//	defer tarfile.Close()
//
//	tr := tar.NewReader(tarfile)
//	var hardlinks []*tar.Header
//	for {
//		hdr, err := tr.Next()
//		if err == io.EOF {
//			// End of tar reached
//			break
//		}
//		if err != nil {
//			return err
//		}
//		switch hdr.Typeflag {
//		case tar.TypeDir:
//			if fileMap != nil {
//				if _, ok := fileMap[hdr.Name]; !ok {
//					continue
//				}
//			}
//			dirpath := path.Join(dst, hdr.Name)
//			ex, err := Exists(dirpath)
//			if err != nil {
//				return err
//			}
//			if ex {
//				err := os.Chmod(dirpath, os.FileMode(hdr.Mode))
//				if err != nil {
//					return err
//				}
//			} else {
//				err := os.MkdirAll(dirpath, os.FileMode(hdr.Mode))
//				if err != nil {
//					return err
//				}
//			}
//		case tar.TypeReg:
//			if fileMap != nil {
//				if _, ok := fileMap[hdr.Name]; !ok {
//					continue
//				}
//			}
//			dir, _ := path.Split(hdr.Name)
//			if dir != "" {
//				err := os.MkdirAll(path.Join(dst, dir), 0755)
//				if err != nil {
//					return err
//				}
//			}
//
//			f, err := os.OpenFile(path.Join(dst, hdr.Name),
//				os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
//			if err != nil {
//				return err
//			}
//			_, err = io.Copy(f, tr)
//			if err != nil {
//				return err
//			}
//			err = f.Close()
//			if err != nil {
//				return err
//			}
//		case tar.TypeSymlink:
//			if fileMap != nil {
//				if _, ok := fileMap[hdr.Name]; !ok {
//					continue
//				}
//			}
//			dir, _ := path.Split(path.Join(dst, hdr.Name))
//			if dir != "" {
//				err := os.MkdirAll(dir, 0755)
//				if err != nil {
//					return err
//				}
//			}
//			err := os.Symlink(hdr.Linkname, path.Join(dst, hdr.Name))
//			if err != nil {
//				return err
//			}
//		case tar.TypeLink:
//			if fileMap != nil {
//				if _, ok := fileMap[hdr.Name]; !ok {
//					continue
//				}
//			}
//			hardlinks = append(hardlinks, hdr)
//		default:
//			return fmt.Errorf("unknown type %c for file in tar: %s",
//				hdr.Typeflag, hdr.Name)
//		}
//	}
//	for _, link := range hardlinks {
//		err := os.Link(link.Linkname, path.Join(dst, link.Name))
//		if err != nil {
//			fmt.Fprintf(os.Stderr, "couldn't create link from %s to %s\n", link.Name, link.Linkname)
//		}
//	}
//	return nil
//}

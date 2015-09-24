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
	"fmt"
	"os"
)

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

func RmIfExists(path string) error {
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
	return nil
}

func RmAndMkdir(path string) error {
	err := RmIfExists(path)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func RmIfPossible(path string, previousError error) error {
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("%v, and couldn't remove %s: %v", previousError,
			path, err)
	}
	return previousError
}

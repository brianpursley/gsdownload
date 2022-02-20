/*
Copyright 2022 Brian Pursley

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var (
	osCreate   = os.Create
	osMkdirAll = os.MkdirAll
	ioCopy     = io.Copy
)

// OsCopier provides the ability to copy data to files
type OsCopier struct {
	mutex sync.Mutex
}

// NewOsCopier creates a new OsCopier instance
func NewOsCopier() *OsCopier {
	return &OsCopier{}
}

// CopyToFile copies data from a reader to a file
func (c *OsCopier) CopyToFile(path string, reader io.Reader) (int64, error) {
	dirPath := filepath.Dir(path)
	err := c.safeMkdirAll(dirPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create directory %s: %v", dirPath, err)
	}

	file, err := osCreate(path)
	if err != nil {
		return 0, fmt.Errorf("failed to create file %s: %v", path, err)
	}
	defer file.Close()

	bytes, err := ioCopy(file, reader)
	if err != nil {
		return 0, fmt.Errorf("failed writing to file %s: %v", path, err)
	}

	return bytes, nil
}

func (c *OsCopier) safeMkdirAll(path string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return osMkdirAll(path, 0755)
}

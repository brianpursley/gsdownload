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
	"io"
)

// MockCopier provides a mock implementation of the Copier interface
type MockCopier struct {
	CopyToFileImplementation func(path string, reader io.Reader) (int64, error)
}

// CopyToFile copies data from a reader to a file
func (c *MockCopier) CopyToFile(path string, reader io.Reader) (int64, error) {
	return c.CopyToFileImplementation(path, reader)
}

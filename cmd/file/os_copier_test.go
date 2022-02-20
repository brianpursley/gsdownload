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
	"bytes"
	"io"
	"os"
	"testing"
)

func TestCopyToFileCreatesAFile(t *testing.T) {
	var createdName string
	osCreate = func(name string) (*os.File, error) {
		createdName = name
		return nil, nil
	}

	var mkdirAllPath string
	var mkdirAllPerm os.FileMode
	osMkdirAll = func(path string, perm os.FileMode) error {
		mkdirAllPath = path
		mkdirAllPerm = perm
		return nil
	}

	var bytesCopied []byte
	ioCopy = func(_ io.Writer, src io.Reader) (written int64, err error) {
		bytesCopied, _ = io.ReadAll(src)
		return int64(len(bytesCopied)), nil
	}

	path := "/foo/bar/baz"
	content := "test content"
	osFileWriter := &OsCopier{}
	byteCount, err := osFileWriter.CopyToFile(path, bytes.NewReader([]byte(content)))
	if err != nil {
		t.Fatal(err)
	}

	expectedMkdirAllPath := "/foo/bar"
	if mkdirAllPath != expectedMkdirAllPath {
		t.Fatalf("wrong mkdirAll path: expected %q, got %q", expectedMkdirAllPath, mkdirAllPath)
	}

	expectedMkdirAllPerm := os.FileMode(0755)
	if mkdirAllPerm != expectedMkdirAllPerm {
		t.Fatalf("wrong mkdirAll perm: expected %v, got %v", expectedMkdirAllPerm, mkdirAllPerm)
	}

	if createdName != path {
		t.Fatalf("wrong create name: expected %q, got %q", path, createdName)
	}

	expectedByteCount := int64(len(content))
	if byteCount != expectedByteCount {
		t.Fatalf("wrong byte count: expected %d, got %d", expectedByteCount, byteCount)
	}

	if string(bytesCopied) != content {
		t.Fatalf("wrong content: expected %q, got %q", content, string(bytesCopied))
	}
}

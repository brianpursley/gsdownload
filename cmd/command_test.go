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

package cmd

import (
	"fmt"
	"github.com/brianpursley/gsdownload/cmd/file"
	"github.com/brianpursley/gsdownload/cmd/storage"
	"io"
	"strconv"
	"sync"
	"testing"
)

func TestConfigureSetsArgs(t *testing.T) {
	runner := runner{}
	err := runner.configure(NewCommand(), []string{"bucket", "prefix", "path"})
	if err != nil {
		t.Fatalf("configure failed: %v", err)
	}
	if runner.bucketName != "bucket" {
		t.Fatalf("wrong bucketName: expected \"bucket\", but got %q", runner.bucketName)
	}
	if runner.prefix != "prefix/" {
		t.Fatalf("wrong prefix: expected \"prefix\", but got %q", runner.prefix)
	}
	if runner.outputDirectory != "path" {
		t.Fatalf("wrong outputDirectory: expected \"path\", but got %q", runner.outputDirectory)
	}
}

func TestConfigurePrefixCases(t *testing.T) {
	testCases := []struct {
		prefix         string
		expectedPrefix string
	}{
		{prefix: "", expectedPrefix: ""},
		{prefix: "/", expectedPrefix: ""},
		{prefix: "foo", expectedPrefix: "foo/"},
		{prefix: "/foo", expectedPrefix: "foo/"},
		{prefix: "/foo/", expectedPrefix: "foo/"},
		{prefix: "foo/", expectedPrefix: "foo/"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Prefix %q should become %q", tc.prefix, tc.expectedPrefix), func(tt *testing.T) {
			runner := runner{}
			err := runner.configure(NewCommand(), []string{"bucket", tc.prefix, "path"})
			if err != nil {
				tt.Fatalf("configure failed: %v", err)
			}
			if runner.prefix != tc.expectedPrefix {
				tt.Fatalf("wrong prefix: expected %q, but got %q", tc.expectedPrefix, runner.prefix)
			}
		})
	}
}

func TestCommandShouldRunSuccessfully(t *testing.T) {
	testObjects := map[string]struct {
		objectInfo   storage.ObjectInfo
		data         []byte
		expectedPath string
	}{
		"prefix/foo": {
			objectInfo:   storage.ObjectInfo{Name: "prefix/foo", Size: 12},
			data:         []byte("prefix/foo contents"),
			expectedPath: "path/foo",
		},
		"prefix/bar": {
			objectInfo:   storage.ObjectInfo{Name: "prefix/bar", Size: 12},
			data:         []byte("prefix/bar contents"),
			expectedPath: "path/bar",
		},
		"prefix/baz/qux": {
			objectInfo:   storage.ObjectInfo{Name: "prefix/baz/qux", Size: 12},
			data:         []byte("prefix/baz/qux contents"),
			expectedPath: "path/baz/qux",
		},
	}

	testCases := map[string]struct {
		dryRun        bool
		maxConcurrent int
		maxObjects    int
		expectError   bool
	}{
		"1 max concurrent, 3 max objects":         {maxConcurrent: 1, maxObjects: 3},
		"1 max concurrent, 2 max objects":         {maxConcurrent: 1, maxObjects: 2, expectError: true},
		"8 max concurrent, 3 max objects":         {maxConcurrent: 8, maxObjects: 3},
		"8 max concurrent, 2 max objects":         {maxConcurrent: 8, maxObjects: 2, expectError: true},
		"dryRun, 8 max concurrent, 3 max objects": {dryRun: true, maxConcurrent: 8, maxObjects: 3},
		"dryRun, 8 max concurrent, 2 max objects": {dryRun: true, maxConcurrent: 8, maxObjects: 2, expectError: true},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			storageClient = &storage.MockClient{
				ObjectInfoProviderFunc: func(bucketName, prefix string) []storage.ObjectInfo {
					if bucketName == "bucket" && prefix == "prefix/" {
						var result []storage.ObjectInfo
						for _, v := range testObjects {
							result = append(result, v.objectInfo)
						}
						return result
					}
					tt.Fatalf("unexpected bucketName/prefix: %s/%s", bucketName, prefix)
					return nil
				},
				ObjectContentProviderFunc: func(bucketName, objectName string) []byte {
					if bucketName == "bucket" {
						if obj, exists := testObjects[objectName]; exists {
							return obj.data
						}
					}
					tt.Fatalf("unexpected bucketName/objectName: %s/%s", bucketName, objectName)
					return nil
				},
			}

			mutex := sync.Mutex{}
			copied := map[string][]byte{}
			fileCopier = &file.MockCopier{
				CopyToFileImplementation: func(path string, reader io.Reader) (int64, error) {
					mutex.Lock()
					defer mutex.Unlock()
					if _, exists := copied[path]; exists {
						return 0, fmt.Errorf("file has already been copied")
					}
					b, _ := io.ReadAll(reader)
					copied[path] = b
					return int64(len(b)), nil
				},
			}

			command := NewCommand()
			command.SetArgs([]string{"bucket", "prefix", "path"})
			_ = command.Flag("max-concurrent").Value.Set(strconv.Itoa(tc.maxConcurrent))
			_ = command.Flag("dry-run").Value.Set(strconv.FormatBool(tc.dryRun))
			_ = command.Flag("max-objects").Value.Set(strconv.Itoa(tc.maxObjects))
			err := command.Execute()
			if tc.expectError {
				if err == nil {
					tt.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				tt.Fatalf("execute failed: %v", err)
			}

			if tc.dryRun {
				if len(copied) != 0 {
					tt.Fatalf("wrong file count: expected 0, got %d", len(copied))
				}
			} else {
				if len(copied) != len(testObjects) {
					tt.Fatalf("wrong file count: expected %d, got %d", len(testObjects), len(copied))
				}
				for _, v := range testObjects {
					expectedData := v.data
					actualData, exists := copied[v.expectedPath]
					if !exists {
						tt.Fatalf("missing file content for path %q", v.expectedPath)
					}
					if string(actualData) != string(expectedData) {
						tt.Fatalf("wrong file contents: expected %q, got %q", expectedData, actualData)
					}
				}
			}
		})
	}
}

func TestCommandShouldHandleNoResults(t *testing.T) {
	testCases := map[string]struct {
		notFoundIsError bool
		expectError     bool
	}{
		"notFoundIsError true should result in an error":      {notFoundIsError: true, expectError: true},
		"notFoundIsError false should not result in an error": {},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			storageClient = &storage.MockClient{
				ObjectInfoProviderFunc: func(bucketName, prefix string) []storage.ObjectInfo {
					return []storage.ObjectInfo{}
				},
				ObjectContentProviderFunc: func(bucketName, objectName string) []byte {
					tt.Fatalf("unexpected call to objectProvider")
					return nil
				},
			}

			fileCopier = &file.MockCopier{
				CopyToFileImplementation: func(path string, reader io.Reader) (int64, error) {
					tt.Fatalf("unexpected call to copyToFileImplementation")
					return 0, nil
				},
			}

			command := NewCommand()
			command.SetArgs([]string{"bucket", "prefix", "path"})
			_ = command.Flag("error").Value.Set(strconv.FormatBool(tc.notFoundIsError))
			err := command.Execute()
			if tc.expectError {
				if err == nil {
					tt.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				tt.Fatalf("execute failed: %v", err)
			}
		})
	}
}

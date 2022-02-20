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

package storage

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
)

// MockClient provides a mock implementation of the Client interface
type MockClient struct {
	connected                 bool
	ObjectInfoProviderFunc    func(bucketName, prefix string) []ObjectInfo
	ObjectContentProviderFunc func(bucketName, objectName string) []byte
}

// Connect simulates a connection being established
func (c *MockClient) Connect(_ context.Context) error {
	c.connected = true
	return nil
}

// Close simulates the connection being closed
func (c *MockClient) Close() error {
	c.connected = false
	return nil
}

// VisitObjects calls a function for each object returned by MockClient.ObjectInfoProviderFunc
func (c *MockClient) VisitObjects(_ context.Context, bucketName, prefix string, visit func(objectInfo ObjectInfo) error) error {
	for _, objectInfo := range c.ObjectInfoProviderFunc(bucketName, prefix) {
		if err := visit(objectInfo); err != nil {
			return err
		}
	}
	return nil
}

// ReadObject returns the data provided by MockClient.ObjectContentProviderFunc
func (c *MockClient) ReadObject(_ context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader(c.ObjectContentProviderFunc(bucketName, objectName))), nil
}

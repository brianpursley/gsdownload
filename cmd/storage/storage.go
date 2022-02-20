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
	"context"
	"io"
)

// Client defines an interface used to interact with Google Cloud Storage
type Client interface {
	Connect(ctx context.Context) error
	VisitObjects(ctx context.Context, bucketName, prefix string, visit func(objectInfo ObjectInfo) error) error
	ReadObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error)
	Close() error
}

// ObjectInfo contains information about an object
type ObjectInfo struct {
	Name string
	Size int64
}

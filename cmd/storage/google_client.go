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
	"fmt"
	"io"
	"os"
	"sync"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GoogleClient provides the ability to interact with Google Cloud Storage
type GoogleClient struct {
	client  *storage.Client
	buckets map[string]*storage.BucketHandle
	mutex   sync.Mutex
}

// NewGoogleClient creates a new instance of GoogleClient
func NewGoogleClient() *GoogleClient {
	return &GoogleClient{
		buckets: map[string]*storage.BucketHandle{},
	}
}

// Connect establishes a connection to Google Cloud Storage
func (c *GoogleClient) Connect(ctx context.Context) error {
	var err error
	c.client, err = storage.NewClient(ctx)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "WARNING: could not find default credentials")
		c.client, err = storage.NewClient(ctx, option.WithoutAuthentication())
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the client
func (c *GoogleClient) Close() error {
	return c.client.Close()
}

func (c *GoogleClient) getBucketHandle(bucketName string) *storage.BucketHandle {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if bucketHandle, exists := c.buckets[bucketName]; exists {
		return bucketHandle
	}
	c.buckets[bucketName] = c.client.Bucket(bucketName)
	return c.buckets[bucketName]
}

// VisitObjects calls a function for each object found in a bucket where the object starts with a specified prefix
func (c *GoogleClient) VisitObjects(ctx context.Context, bucketName, prefix string, visit func(objectInfo ObjectInfo) error) error {
	bucket := c.getBucketHandle(bucketName)
	query := &storage.Query{Prefix: prefix}
	err := query.SetAttrSelection([]string{"Name", "Size"})
	if err != nil {
		return err
	}
	objectIterator := bucket.Objects(ctx, query)
	for {
		objAttrs, err := objectIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		objectInfo := ObjectInfo{
			Name: objAttrs.Name,
			Size: objAttrs.Size,
		}
		if err := visit(objectInfo); err != nil {
			return err
		}
	}
	return nil
}

// ReadObject reads the content of an object from Google Cloud Storage
func (c *GoogleClient) ReadObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	bucket := c.getBucketHandle(bucketName)
	return bucket.Object(objectName).NewReader(ctx)
}

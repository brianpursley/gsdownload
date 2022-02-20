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
	"context"
	"fmt"
	"github.com/brianpursley/gsdownload/cmd/file"
	"github.com/brianpursley/gsdownload/version"
	"path/filepath"
	"strings"

	"github.com/brianpursley/gsdownload/cmd/storage"
	"github.com/spf13/cobra"
)

var (
	storageClient storage.Client = storage.NewGoogleClient()
	fileCopier    file.Copier    = file.NewOsCopier()
)

type runner struct {
	bucketName      string
	prefix          string
	outputDirectory string

	dryRun          bool
	notFoundIsError bool
	maxConcurrent   int
	maxObjects      int
	verbose         bool
	version         bool
}

// NewCommand creates a new instance of the command
func NewCommand() *cobra.Command {
	r := runner{}

	var cmd = &cobra.Command{
		Use:          "gsdownload <bucket> <prefix> <output directory>",
		Short:        "Bulk download objects from a Google Cloud Storage bucket",
		Long:         `A utility for downloading objects from a Google Cloud Storage bucket`,
		SilenceUsage: true,
		RunE:         r.run,
	}

	cmd.Flags().BoolVar(&r.dryRun, "dry-run", false, "Display a list of the files that will be downloaded and then exit without downloading them")
	cmd.Flags().IntVar(&r.maxConcurrent, "max-concurrent", 8, "The maximum number of concurrent downloads (0=unlimited)")
	cmd.Flags().IntVar(&r.maxObjects, "max-objects", 1000, "The maximum number of objects to download (0=unlimited)")
	cmd.Flags().BoolVar(&r.notFoundIsError, "error", false, "Exit with non-zero exit code if no objects were found matching the specified prefix")
	cmd.Flags().BoolVarP(&r.verbose, "verbose", "v", false, "Include additional information about each object that is downloaded")
	cmd.Flags().BoolVar(&r.version, "version", false, "Print version information and exit")

	return cmd
}

func (r *runner) configure(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(3)(cmd, args); err != nil {
		return err
	}

	r.bucketName = args[0]
	r.prefix = args[1]
	r.outputDirectory = args[2]

	if r.maxConcurrent < 0 {
		return fmt.Errorf("--max-concurrent must be greater than or equal to zero")
	}

	if r.maxObjects < 0 {
		return fmt.Errorf("--max-objects must be greater than or equal to zero")
	}

	if !strings.HasSuffix(r.prefix, "/") {
		r.prefix = r.prefix + "/"
	}
	if r.prefix == "/" {
		r.prefix = ""
	}
	r.prefix = strings.TrimPrefix(r.prefix, "/")

	return nil
}

func (r *runner) run(cmd *cobra.Command, args []string) error {
	if r.version {
		fmt.Println(version.Version)
		return nil
	}

	if err := r.configure(cmd, args); err != nil {
		return err
	}

	err := storageClient.Connect(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}
	defer storageClient.Close()

	objects, err := r.getObjects(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get objects: %v", err)
	}

	if r.notFoundIsError && len(objects) == 0 {
		return fmt.Errorf("no objects found")
	}

	var sem chan bool
	if r.maxConcurrent > 0 {
		sem = make(chan bool, r.maxConcurrent)
	}

	errorChan := make(chan error)
	go func() {
		for _, obj := range objects {
			if sem != nil {
				sem <- true
			}
			go func(obj *storage.ObjectInfo) {
				if sem != nil {
					defer func() { <-sem }()
				}
				if r.dryRun {
					r.printObject(obj.Name, obj.Size)
					errorChan <- nil
				} else {
					errorChan <- r.downloadObject(cmd.Context(), obj.Name)
				}
			}(obj)
		}
	}()

	for range objects {
		err := <-errorChan
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *runner) getObjects(ctx context.Context) ([]*storage.ObjectInfo, error) {
	var objects []*storage.ObjectInfo
	err := storageClient.VisitObjects(ctx, r.bucketName, r.prefix, func(objectInfo storage.ObjectInfo) error {
		if strings.HasSuffix(objectInfo.Name, "/") {
			// Skip directories
			return nil
		}
		objects = append(objects, &objectInfo)
		if r.maxObjects > 0 && len(objects) > r.maxObjects {
			return fmt.Errorf("exceeded the maximum number of objects")
		}
		return nil
	})
	return objects, err
}

func (r *runner) downloadObject(ctx context.Context, objectName string) error {
	reader, err := storageClient.ReadObject(ctx, r.bucketName, objectName)
	if err != nil {
		return fmt.Errorf("failed to create new reader for %s: %v", objectName, err)
	}
	defer reader.Close()

	path := r.getPathForObject(objectName)
	bytes, err := fileCopier.CopyToFile(path, reader)
	if err != nil {
		return fmt.Errorf("failed writing to file %s: %v", objectName, err)
	}

	r.printObject(objectName, bytes)
	return nil
}

func (r *runner) getPathForObject(name string) string {
	nameWithoutPrefix := strings.TrimPrefix(name, r.prefix)
	return filepath.Join(r.outputDirectory, nameWithoutPrefix)
}

func (r *runner) printObject(name string, size int64) {
	if r.verbose {
		fmt.Printf("%s --> %s (size=%d)\n", name, r.getPathForObject(name), size)
	} else {
		fmt.Println(name)
	}
}

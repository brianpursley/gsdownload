# gsdownload

gsdownload is a utility for bulk downloading multiple objects from a Google Cloud Storage bucket.

## Usage

```
A utility for downloading objects from a Google Cloud Storage bucket

Usage:
  gsdownload <bucket> <prefix> <output directory> [flags]

Flags:
      --dry-run              Display a list of the files that will be downloaded and then exit without downloading them
      --error                Exit with non-zero exit code if no objects were found matching the specified prefix
  -h, --help                 help for gsdownload
      --max-concurrent int   The maximum number of concurrent downloads (0=unlimited) (default 8)
      --max-objects int      The maximum number of objects to download (0=unlimited) (default 1000)
  -v, --verbose              Include additional information about each object that is downloaded
      --version              Print version information and exit
```

### Examples

#### Download all objects from the `foo` bucket that start with `/bar/baz` prefix and save them in the local `/tmp/objects` directory
```
gsdownload foo /bar/baz /tmp/objects
```

#### Download all objects from the `foo` bucket and save them in the current directory
```
gsdownlaoad foo / .
```

## Building from source

Install tool dependencies.
```
go install honnef.co/go/tools/cmd/staticcheck@latest
go install golang.org/x/lint/golint@latest
```

Run make.
```
make
```

Optionally, target a specific OS and Architecture.
```
GOOS=windows GOARCH=amd64 make build
```

Build output is generated in the `_output` directory.

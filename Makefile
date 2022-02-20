# Copyright 2022 Brian Pursley
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ifeq ("${shell go env GOOS}","windows")
	OUTPUT ?= _output/gsdownload.exe
else
	OUTPUT ?= _output/gsdownload
endif

.PHONY: all
all: staticcheck lint vet verify test build

.PHONY: staticcheck
staticcheck:
ifeq (, $(shell which staticcheck))
	$(error staticcheck not found (go install honnef.co/go/tools/cmd/staticcheck@latest))
endif
	staticcheck ./...

.PHONY: lint
lint:
ifeq (, $(shell which golint))
	$(error golint not found (go install golang.org/x/lint/golint@latest))
endif
	golint -set_exit_status ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: verify
verify:
	go mod verify

.PHONY: test
test:
	go test ./...

.PHONY: testrace
testrace:
	go test -race ./...

.PHONY: build
build:
	go build \
		-ldflags "-s -w -X github.com/brianpursley/gsdownload/version.Version=${shell git describe --tags --dirty --match "v[0-9].[0-9].[0-9]"}" \
		-o $(OUTPUT)

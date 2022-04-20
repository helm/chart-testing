#!/usr/bin/env bash

# Copyright The Helm Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit

go install github.com/goreleaser/goreleaser@v1.8.2
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/sigstore/cosign/cmd/cosign@v1.7.2
go install github.com/anchore/syft@v0.44.1

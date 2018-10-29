# Copyright 2018 The Helm Authors. All rights reserved.
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

FROM bash:4.4

RUN apk --no-cache add \
    curl \
    git \
    jq \
    libc6-compat \
    openssh-client \
    python \
    py-crcmod \
    py-pip

# Install YQ command line reader
ARG YQ_VERSION=2.7.0
RUN pip install "yq==$YQ_VERSION"

# Install SemVer testing tool
ARG VERT_VERSION=0.1.0
RUN curl -Lo vert "https://github.com/Masterminds/vert/releases/download/v$VERT_VERSION/vert-v$VERT_VERSION-linux-amd64" && \
    chmod +x vert && \
    mv vert /usr/local/bin/

# Install a YAML Linter
ARG YAML_LINT_VERSION=1.8.1
RUN pip install "yamllint==$YAML_LINT_VERSION"

# Install Yamale YAML schema validator
ARG YAMALE_VERSION=1.7.0
RUN pip install "yamale==$YAMALE_VERSION"

# Install kubectl
ARG KUBECTL_VERSION=1.12.2
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/v$KUBECTL_VERSION/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install Helm
ARG HELM_VERSION=2.11.0
RUN curl -LO "https://kubernetes-helm.storage.googleapis.com/helm-v$HELM_VERSION-linux-amd64.tar.gz" && \
    mkdir -p "/usr/local/helm-$HELM_VERSION" && \
    tar -xzf "helm-v$HELM_VERSION-linux-amd64.tar.gz" -C "/usr/local/helm-$HELM_VERSION" && \
    ln -s "/usr/local/helm-$HELM_VERSION/linux-amd64/helm" /usr/local/bin/helm && \
    rm -f "helm-v$HELM_VERSION-linux-amd64.tar.gz"

COPY etc /testing/etc/
COPY lib /testing/lib/
COPY chart_test.sh /testing/

RUN ln -s /testing/chart_test.sh /usr/local/bin/chart_test.sh


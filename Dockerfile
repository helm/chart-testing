FROM alpine:3.8

RUN apk --no-cache add \
    curl \
    git \
    libc6-compat \
    openssh-client \
    python \
    py-crcmod \
    py-pip

# Install a YAML Linter
ARG YAML_LINT_VERSION=1.8.1
RUN pip install "yamllint==$YAML_LINT_VERSION"

# Install Yamale YAML schema validator
ARG YAMALE_VERSION=1.7.0
RUN pip install "yamale==$YAMALE_VERSION"

# Install kubectl
ARG KUBECTL_VERSION=v1.12.0
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install Helm
ARG HELM_VERSION=v2.11.0
RUN curl -LO "https://kubernetes-helm.storage.googleapis.com/helm-$HELM_VERSION-linux-amd64.tar.gz" && \
    mkdir -p "/usr/local/helm-$HELM_VERSION" && \
    tar -xzf "helm-$HELM_VERSION-linux-amd64.tar.gz" -C "/usr/local/helm-$HELM_VERSION" && \
    ln -s "/usr/local/helm-$HELM_VERSION/linux-amd64/helm" /usr/local/bin/helm && \
    rm -f "helm-$HELM_VERSION-linux-amd64.tar.gz"

# Goreleaser needs to override this because it builds the
# Dockerfile from a tmp dir with all files to be copied in the root
ARG dist_dir=dist/linux_amd64

COPY "$dist_dir/chart_schema.yaml" /etc/ct/chart_schema.yaml
COPY "$dist_dir/lintconf.yaml" /etc/ct/lintconf.yaml
COPY "$dist_dir/ct" /usr/local/bin/ct

FROM alpine:3.9

RUN apk --no-cache add \
    curl \
    git \
    libc6-compat \
    openssh-client \
    python \
    py-crcmod \
    py-pip && \
    pip install --upgrade pip==18.1

# Install a YAML Linter
ARG YAML_LINT_VERSION=1.13.0
RUN pip install "yamllint==$YAML_LINT_VERSION"

# Install Yamale YAML schema validator
ARG YAMALE_VERSION=1.8.0
RUN pip install "yamale==$YAMALE_VERSION"

# Install kubectl
ARG KUBECTL_VERSION=v1.14.1
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install Helm
ARG HELM_VERSION=v2.13.1
RUN curl -LO "https://kubernetes-helm.storage.googleapis.com/helm-$HELM_VERSION-linux-amd64.tar.gz" && \
    mkdir -p "/usr/local/helm-$HELM_VERSION" && \
    tar -xzf "helm-$HELM_VERSION-linux-amd64.tar.gz" -C "/usr/local/helm-$HELM_VERSION" && \
    ln -s "/usr/local/helm-$HELM_VERSION/linux-amd64/helm" /usr/local/bin/helm && \
    rm -f "helm-$HELM_VERSION-linux-amd64.tar.gz"

COPY ./etc/chart_schema.yaml /etc/ct/chart_schema.yaml
COPY ./etc/lintconf.yaml /etc/ct/lintconf.yaml
COPY ct /usr/local/bin/ct
# Ensure that the binary is available on path and is executable
RUN ct --help

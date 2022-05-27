FROM alpine:3.15

RUN apk --no-cache add \
    bash \
    curl>7.77.0-r0 \
    git \
    libc6-compat \
    openssh-client \
    py3-pip \
    py3-wheel \
    python3 && \
    pip install --upgrade pip==21.0.1

# Install a YAML Linter
ARG yamllint_version=1.26.3
LABEL yamllint_version=$yamllint_version
RUN pip install "yamllint==$yamllint_version"

# Install Yamale YAML schema validator
ARG yamale_version=4.0.0
LABEL yamale_version=$yamale_version
RUN pip install "yamale==$yamale_version"

ARG TARGETPLATFORM
# Install kubectl
ARG kubectl_version=v1.23.6
LABEL kubectl_version=$kubectl_version
RUN curl -LO "https://storage.googleapis.com/kubernetes-release/release/$kubectl_version/bin/$TARGETPLATFORM/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install Helm
ARG helm_version=v3.9.0
LABEL helm_version=$helm_version
RUN targetArch=$(echo $TARGETPLATFORM | cut -f2 -d '/') \
    && if [ ${targetArch} = "amd64" ]; then \
    HELM_ARCH="linux-amd64"; \
elif [ ${targetArch} = "arm64" ]; then \
    HELM_ARCH="linux-arm64"; \
fi \
    && curl -LO "https://get.helm.sh/helm-$helm_version-$HELM_ARCH.tar.gz" \
    && mkdir -p "/usr/local/helm-$helm_version" \
    && tar -xzf "helm-$helm_version-$HELM_ARCH.tar.gz" -C "/usr/local/helm-$helm_version" \
    && ln -s "/usr/local/helm-$helm_version/$HELM_ARCH/helm" /usr/local/bin/helm \
    && rm -f "helm-$helm_version-$HELM_ARCH.tar.gz"

COPY ./etc/chart_schema.yaml /etc/ct/chart_schema.yaml
COPY ./etc/lintconf.yaml /etc/ct/lintconf.yaml
COPY ct /usr/local/bin/ct
# Ensure that the binary is available on path and is executable
RUN ct --help

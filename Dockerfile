FROM alpine:3.22

RUN apk --no-cache add \
    bash \
    curl \
    git \
    libc6-compat \
    openssh-client \
    py3-pip \
    py3-wheel \
    python3 \
    yamllint

# Install Yamale YAML schema validator
# Note: Yamale 6.0.0 supports Python 3.8+, including Python 3.13.
# Python 3.14 compatibility requires Yamale >= 7.0.0 (when available).
# For Python 3.14 runners, ensure Yamale is updated to a version with Python 3.14 support.
ARG yamale_version=6.0.0
LABEL yamale-version=$yamale_version
RUN pip install --break-system-packages "yamale==$yamale_version"

ARG TARGETPLATFORM
# Install kubectl
ARG kubectl_version=v1.32.0
LABEL kubectl-version=$kubectl_version
RUN targetArch=$(echo $TARGETPLATFORM | cut -f2 -d '/') \
    && if [ ${targetArch} = "amd64" ]; then \
    HELM_ARCH="linux/amd64"; \
elif [ ${targetArch} = "arm64" ]; then \
    HELM_ARCH="linux/arm64"; \
fi \
    && curl -LO "dl.k8s.io/$kubectl_version/bin/$HELM_ARCH/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/

# Install Helm
ARG helm_version=v3.16.4
LABEL helm-version=$helm_version
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
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/ct /usr/local/bin/ct
# Ensure that the binary is available on path and is executable
RUN ct --help

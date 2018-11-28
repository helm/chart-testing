## BUILDER
#
FROM alpine:3.8 as build

ARG KUBECTL_VERSION=v1.12.2
ARG HELM_VERSION=v2.11.0

WORKDIR /artifacts

# Goreleaser needs to override this because it builds the
# Dockerfile from a tmp dir with all files to be copied in the root
ARG dist_dir=dist/linux_amd64

COPY "$dist_dir/chart_schema.yaml"  "$dist_dir/lintconf.yaml" /etc/ct/
COPY "$dist_dir/ct" /usr/local/bin/

RUN apk --no-cache add curl && \
    curl -sLo /artifacts/kubectl "https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    curl -sLo /var/tmp/helm.tgz "https://kubernetes-helm.storage.googleapis.com/helm-$HELM_VERSION-linux-amd64.tar.gz" && \
    tar -xzf /var/tmp/helm.tgz -C /var/tmp && \
    mv /var/tmp/linux-amd64/helm /artifacts

## WORKER
#
FROM python:3.7.1-alpine3.8

ARG YAML_LINT_VERSION=1.12.1
ARG YAMALE_VERSION=1.7.0

RUN pip install \
    "yamllint==$YAML_LINT_VERSION" \
    "yamale==$YAMALE_VERSION" && \
    rm -rf /lib/apk/ /etc/apk/ /root/cache/

COPY --from=build /artifacts/* /usr/local/bin/
COPY --from=build /etc/ct/* /etc/ct/

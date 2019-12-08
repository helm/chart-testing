FROM alpine:3.10

ARG HELM_VERSION=v3.0.1
ARG KUBECTL_VERSION=v1.16.2
ARG YAMALE_VERSION=2.0.1
ARG YAML_LINT_VERSION=1.19.0

COPY ./etc/chart_schema.yaml /etc/ct/chart_schema.yaml
COPY ./etc/lintconf.yaml /etc/ct/lintconf.yaml
COPY ct /usr/local/bin/ct
COPY docker-install.sh /tmp/docker-install.sh

RUN chmod +x /tmp/docker-install.sh && \
    /bin/sh -c /tmp/docker-install.sh && \
    rm /tmp/*

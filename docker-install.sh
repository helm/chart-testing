#!/bin/sh

set -o errexit
set -o nounset

# Install apk packages
apk --no-cache add curl git libc6-compat openssh-client python py-crcmod py-pip 

# Upgrade pip
pip install --upgrade pip==18.1

# Install a YAML Linter
pip install "yamllint==$YAML_LINT_VERSION"

# Install Yamale YAML schema validator
pip install "yamale==$YAMALE_VERSION"

# Install kubectl
curl --silent --show-error --fail --location --output /usr/local/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/linux/amd64/kubectl"
chmod +x /usr/local/bin/kubectl

# Install Helm
curl -LO "https://get.helm.sh/helm-$HELM_VERSION-linux-amd64.tar.gz"
mkdir -p "/usr/local/helm-$HELM_VERSION"
tar -xzf "helm-$HELM_VERSION-linux-amd64.tar.gz" -C "/usr/local/helm-$HELM_VERSION"
ln -s "/usr/local/helm-$HELM_VERSION/linux-amd64/helm" /usr/local/bin/helm
rm -f "helm-$HELM_VERSION-linux-amd64.tar.gz"

# # Ensure that the binary is available on path and is executable
ct --help

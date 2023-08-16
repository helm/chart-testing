#!/usr/bin/env bash

# What is this script doing?

# user setup for accessing mounted files
# mount required volumes (they are RW because ct writes files to disk)
# set HOME so it isn't /
# include extra docker args as needed
# ridiculous bash magic to extend PATH within the container shell


# mount special files if they exist
if [[ -d "${GCLOUD_SDK_PATH:-/usr/lib/google-cloud-sdk}" ]]; then
    # include gcloud SDK if available for accessing Google Artifact Registry
    DOCKER_ARGS="${DOCKER_ARGS} --volume ${GCLOUD_SDK_PATH:-/usr/lib/google-cloud-sdk}:/usr/lib/google-cloud-sdk:ro"
    # extend the PATH variable to include the gcloud SDK binaries
    EXTEND_PATH='export PATH=$PATH:/usr/lib/google-cloud-sdk/bin;'
fi

# run the container image using docker
docker run --rm -t \
    -u $(id -u):$(id -g) \
    --volume  ${HOME:?}:/tmp/home:rw,Z  \
    --volume  $(pwd):/src:rw,Z \
    --workdir /src \
    -e HOME=/tmp/home \
    ${DOCKER_ARGS} \
    ${IMAGE_NAME:-quay.io/helmpack/chart-testing}:${IMAGE_TAG:-latest} \
    /bin/bash -c 'echo $@ | bash -v' -- ${EXTEND_PATH} ${@}

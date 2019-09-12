# Chart testing example with Google Cloud Build

This example shows how to lint and test charts using [Google Cloud Build](https://cloud.google.com/cloud-build/)

Since Google Cloud Build will ignore copying over `.git` by default, you will need to initialize `git` and add a `remote`. This example assumes that there is a pre-existing GKE cluster with `helm` already installed.

# Chart testing example with CircleCi and kind - `K`ubernetes `in` `D`ocker

`kind` is a tool for running local Kubernetes clusters using Docker container "nodes".

This example shows how to lint and test charts using CircleCi and [kind](https://github.com/kubernetes-sigs/kind).
It creates a cluster with a single control-plane node and one worker node.
The cluster configuration can be adjusted in [kind-config.yaml](test/kind-config.yaml). You can check for available configuration options in `kind` [docs](https://kind.sigs.k8s.io/docs/user/quick-start#configuring-your-kind-cluster)

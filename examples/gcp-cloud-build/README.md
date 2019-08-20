# Chart testing example with Google Cloud Build

Since Google Cloud Build will ignore copying over `.git` by default, you will need to `git init` and `git remote add`.
Please see `cloudbuild.yaml` for an example on how to lint charts using Google Cloud build.

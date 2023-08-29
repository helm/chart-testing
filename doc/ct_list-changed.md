## ct list-changed

List changed charts

### Synopsis

"List changed charts based on configured charts directories,
"remote, and target branch

```
ct list-changed [flags]
```

### Options

```
      --chart-dirs strings        Directories containing Helm charts. May be specified multiple times
                                  or separate values with commas (default [charts])
      --config string             Config file
      --exclude-deprecated        Skip charts that are marked as deprecated
      --excluded-charts strings   Charts that should be skipped. May be specified multiple times
                                  or separate values with commas
      --github-groups             Change the delimiters for github to create collapsible groups
                                  for command output
  -h, --help                      help for list-changed
      --print-config              Prints the configuration to stderr (caution: setting this may
                                  expose sensitive data when helm-repo-extra-args contains passwords)
      --remote string             The name of the Git remote used to identify changed charts (default "origin")
      --since string              The Git reference used to identify changed charts (default "HEAD")
      --target-branch string      The name of the target branch used to identify changed charts (default "main")
      --use-helmignore            Use .helmignore when identifying changed charts
```

### SEE ALSO

* [ct](ct.md)	 - The Helm chart testing tool

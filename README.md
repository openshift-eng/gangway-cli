# Gangway CLI

A really simple quick CLI client for the gangway API. Needs some polish,
should probably handle reading the response back and calling the status
API.

Set MY_APPICI_TOKEN to your cluster token, and then build and call the
CLI.

Example:


```
$ export MY_APPCI_TOKEN="sha256~secret"
$ go build .
$ ./gangway-cli \
    --api-url="https://gangway-ci.apps.somecluster.org" \
    --initial "registry.ci.openshift.org/ocp/release:4.14.0-0.nightly-2023-07-05-071214" \
    --latest "registry.ci.openshift.org/ocp/release:4.14.0-0.nightly-2023-07-05-071214" \
    --job-name periodic-ci-openshift-release-master-ci-4.14-e2e-aws-ovn
```

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


Use --jobs-file-path to capture JobRunIdentifier JSON which can be fed into other tooling like aggregation analysis

```
$ ./gangway-cli \
    --api-url="https://gangway-ci.apps.somecluster.org" \
    --initial "registry.ci.openshift.org/ocp/release:4.14.0-0.nightly-2023-07-05-071214" \
    --latest "registry.ci.openshift.org/ocp/release:4.14.0-0.nightly-2023-07-05-071214" \
    --job-name periodic-ci-openshift-release-master-ci-4.15-upgrade-from-stable-4.14-e2e-aws-ovn-upgrade \
    --n 10 \
    --jobs-file-path="/path/to/results/for/gangway/"
```

Example input to job-run-aggregator

```
 ./job-run-aggregator analyze-job-runs \
  --timeout=7h \
   --working-dir /path/to/results/for/testaggregation \
   --google-service-account-credential-file /path/to/gcp/credentials.json \
   --job periodic-ci-openshift-release-master-ci-4.15-upgrade-from-stable-4.14-e2e-aws-ovn-upgrade \
   --payload-tag GANGWAY \
   --job-start-time 2023-10-18T13:20:13Z \ 
   --static-run-info-path=/path/to/results/for/gangway/gangway_periodic-ci-openshift-release-master-ci-4.15-upgrade-from-stable-4.14-e2e-aws-ovn-upgrade_2023-10-18T13:20:13.15912424-04:00.json

```
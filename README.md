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
    --num-jobs 10 \
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

More info: https://docs.ci.openshift.org/docs/how-tos/triggering-prowjobs-via-rest/

## FAQ

## How to Get the Token

- Log in a registry, i.e: https://oauth-openshift.apps.ci.l2s4.p1.openshiftapps.com/oauth/token/display
- Click your username in the top-right corner of the window
- Click "Copy login command"
- Click "Display token"
- Copy the value under "Your API token is"

For further information, see: https://docs.ci.openshift.org/docs/how-tos/use-registries-in-build-farm/#summary-of-available-registries

## What images to use?

You might want to check a build from: https://amd64.ocp.releases.ci.openshift.org/

`--initial` only needs to be different from `--latest` if you are testing an upgrade (usually not needed)

## How to Find the Jobs

You might find it easier to find the job names you want to test from Sippy:
https://sippy.dptools.openshift.org/sippy-ng/jobs


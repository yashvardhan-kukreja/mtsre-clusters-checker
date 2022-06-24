# MTSRE Clusters Checker

A friendly helper which scans for MTSRE clusters and sees if anyone of them is worth cleaning up or not to save some bucks :)

## Execution

### Prerequisities

The following credentials are required to run the MTSRE Clusters Checker:
* `OCM_TOKEN`: OCM Token which allows you to read the MTSRE clusters.
* `SLACK_TOKEN`: Slack Token which is authorized to post message on MTSRE's Slack handle.

### Running the MTSRE Clusters Checker

* Have Go 1.17 on your host.
* Run the following:
```sh
make run
```

### Cleanup

* Run the following:
```sh
make clean
```

## Release and Deployment

The deployment of this tool is GitOps-driven over the internal Red Hat's app-interface pipelines.

Those pipelines recognise a change under `./deploy/cronjob-template.yaml` and accordingly, re-deploy the cronjob.

Hence, to cut a new release against this whole process, raise a PR to update the CronJob's `image` with the latest image pushed to Quay registry corresponding to `quay.io/mtsre/mtsre-clusters-checker`.

## Contributions

Feel free to suggest anything by raising an Issue on this repository, getting a clearance from the maintainers and proceeding to work on its associated PR.

## Future? 

Carve this into a CLI which can be used by any team in the future to perform the same kind of checks for the Openshift clusters belonging to their org.
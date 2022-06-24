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

## Contributions

Feel free to suggest anything by raising an Issue on this repository, getting a clearance and proceeding to work on its associated PR.

## Future? 

Carve this into a CLI which can be used by any team in the future to perform the same kind of checks for the Openshift clusters belonging to their org.
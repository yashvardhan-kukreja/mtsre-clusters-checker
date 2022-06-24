package api

import (
	accountsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	clustersv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type ClusterInstance struct {
	Cluster clustersv1.Cluster
	Account accountsv1.Account
}

type ClusterCheckupResult struct {
	Success string
	Failure string
	Error   error
}

type Environment struct {
	Alias      string
	Url        string
	MtSreOrgId string
}

var (
	Production Environment = Environment{
		Alias:      "production",
		Url:        "https://api.openshift.com",
		MtSreOrgId: "1u1wd0UgiYs6ia6RtNedCgq5bB2",
	}
	Staging Environment = Environment{
		Alias:      "staging",
		Url:        "https://api.stage.openshift.com",
		MtSreOrgId: "1u1wPVuTr2m2yyP59aG1qQlvA3D",
	}
	Integration Environment = Environment{
		Alias:      "integration",
		Url:        "https://api.integration.openshift.com",
		MtSreOrgId: "1vZurUnxOx3tHqa6ecj9OWAYy6P",
	}
)

const MTSRE_INFO_CHANNEL_ID = "C01V4S8GXPD"

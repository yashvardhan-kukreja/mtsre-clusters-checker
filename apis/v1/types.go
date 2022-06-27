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
	Alias string
	Url   string
	OrgId string
}

var (
	Production Environment = Environment{
		Alias: "production",
		Url:   "https://api.openshift.com",
	}
	Staging Environment = Environment{
		Alias: "staging",
		Url:   "https://api.stage.openshift.com",
	}
	Integration Environment = Environment{
		Alias: "integration",
		Url:   "https://api.integration.openshift.com",
	}
)

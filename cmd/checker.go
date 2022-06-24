package main

import (
	"log"
	"os"

	apisv1 "github.com/mt-sre/mtsre-clusters-checker/apis/v1"
	"github.com/mt-sre/mtsre-clusters-checker/internal"
)

func main() {
	ocmToken, found := os.LookupEnv("OCM_TOKEN")
	if !found {
		log.Fatal("OCM_TOKEN environment variable not found!")
	}

	envs := []apisv1.Environment{apisv1.Integration, apisv1.Staging, apisv1.Production}

	var consolidatedMsg string
	for _, env := range envs {
		result := internal.PerformClustersCheckup(ocmToken, env)
		if result.Success != "" {
			consolidatedMsg += result.Success + "\n"
		}
		if result.Failure != "" {
			consolidatedMsg += result.Failure + "\n"
		}
		if result.Error != nil {
			consolidatedMsg += result.Error.Error() + "\n"
		}
		consolidatedMsg += "---------------------\n\n"
	}

	if err := internal.NotifyOnSlack(apisv1.MTSRE_INFO_CHANNEL_ID, consolidatedMsg); err != nil {
		log.Fatal("failed to notify about MTSRE's stale cluster instances on slack", err)
	}
	log.Println("MTSRE Cluster Checkup performed successfully")
}

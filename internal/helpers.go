package internal

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"encoding/json"

	apisv1 "github.com/mt-sre/mtsre-clusters-checker/apis/v1"
	"github.com/openshift-online/ocm-cli/pkg/ocm"
	sdk "github.com/openshift-online/ocm-sdk-go"
	accountsv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	clustersv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func FetchStaleClusterInstances(connection *sdk.Connection, orgId string) ([]apisv1.ClusterInstance, error) {
	request := connection.ClustersMgmt().V1().Clusters().List()
	// Send the request till we receive a page with less items than requested:
	size := 100
	page := 1
	targets := []apisv1.ClusterInstance{}
	for {
		// Fetch the next page:
		request.Size(size)
		request.Page(page)
		response, err := request.Send()
		if err != nil {
			return targets, fmt.Errorf("failed to list clusters for the page %d: %w", page, err)
		}

		processCluster := func(item *clustersv1.Cluster, wg *sync.WaitGroup) {
			defer wg.Done()

			clusterId := item.ID()
			log.Default().Printf("Evaluating the cluster: %s...", clusterId)
			creationTimestamp, ok := item.GetCreationTimestamp()
			if !ok {
				return
			}

			// ignore the clusters which are younger than 24 hours
			if time.Since(creationTimestamp) <= 24*time.Hour {
				log.Default().Printf("Skipping the check for the cluster %s as it's younger than 24hr", clusterId)
				return
			}

			sub, ok := item.GetSubscription()
			if !ok {
				log.Default().Printf("Unable to fetch the subscription metadata of the cluster %s. Skipping its check", clusterId)
				return
			}

			subId, ok := sub.GetID()
			if !ok {
				log.Default().Printf("Unable to fetch the subscription id of the cluster %s. Skipping its check", clusterId)
				return
			}

			subResp, err := connection.AccountsMgmt().V1().Subscriptions().Subscription(subId).Get().Send()
			if err != nil {
				log.Default().Printf("Unable to fetch the subscription details associated with the subscription %s of the cluster %s. Skipping its check", subId, clusterId)
				return
			}

			respOrgId, ok := subResp.Body().GetOrganizationID()
			if !ok {
				log.Default().Printf("Unable to fetch the organization id from the subscription %s of the cluster %s. Skipping its check", subId, clusterId)
				return
			}

			if respOrgId == orgId {
				log.Default().Printf("The Cluster %s found to be older than 24h!", clusterId)
				account := fetchAccountDetails(connection, subResp.Body().Creator().ID())
				if reflect.DeepEqual(account, accountsv1.Account{}) {
					log.Default().Printf("Unable to find the account details of the owner of the cluster %s. Skipping the account details and proceeding further.", clusterId)
				}
				targets = append(targets, apisv1.ClusterInstance{
					Cluster: *item,
					Account: account,
				})
			}
		}

		wg := &sync.WaitGroup{}
		wg.Add(response.Items().Len())

		// TODO(optional): Create a worker pool of the size of the number of cores available. This will achieve parallelism without the context switching overhead of blind concurrency like happening below
		response.Items().Each(func(item *clustersv1.Cluster) bool {
			go processCluster(item, wg)
			return true
		})
		wg.Wait()

		// If the number of fetched items is less than requested, then this was the last
		// page, otherwise process the next one:
		if response.Size() < size {
			break
		}
		page++
	}
	return targets, nil
}

func fetchAccountDetails(connection *sdk.Connection, accountId string) accountsv1.Account {
	if accountId == "" {
		log.Default().Printf("Account ID found to be empty")
		return accountsv1.Account{}
	}
	accountResp, err := connection.AccountsMgmt().V1().Accounts().Account(accountId).Get().Send()
	if err != nil {
		log.Default().Printf("Failed to get the account details corresponding to the account id %s: %v", accountId, err.Error())
		return accountsv1.Account{}
	}
	return *accountResp.Body()
}

func GenerateNotificationMessage(clusterInstances []apisv1.ClusterInstance, env apisv1.Environment) string {
	message := fmt.Sprintf("Clusters getting checked for any stale clusters in *%s* environment...\n", env.Alias)
	if len(clusterInstances) == 0 {
		message += "No clusters were found to be older than 24h :D"
		return message
	}
	message += "The following active clusters were found to be older than 24h. Please remove them if they aren't required anymore:\n"
	for _, clusterInstance := range clusterInstances {
		message += fmt.Sprintf(`
*Cluster*: %s
*ID*: %s
*Owner*: %s 
*Age / Creation Timestamp*: %s / %s
*Environment*: %s
`, clusterInstance.Cluster.Name(), clusterInstance.Cluster.ExternalID(), clusterInstance.Account.Username(), time.Since(clusterInstance.Cluster.CreationTimestamp()), clusterInstance.Cluster.CreationTimestamp(), env.Alias)
	}
	return strings.TrimSuffix(message, "\n")
}

func NotifyOnSlack(token, channel, message string) error {
	data := map[string]string{"channel": channel, "text": message, "parse": "full"}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal slack request data: %w", err)
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to generate the POST request to Slack API for sending the message: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute the POST request to Slack API for sending the message: %w", err)
	}
	defer resp.Body.Close()

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("failed to unmarshal the response from Slack API: %w", err)
	}

	success, ok := res["ok"].(bool)
	if !success || !ok {
		return fmt.Errorf("the Slack API request didn't succeed: %+v", resp.Body)
	}

	return nil
}

func PerformClustersCheckup(ocmToken string, env apisv1.Environment) apisv1.ClusterCheckupResult {
	log.Printf("Performing clusters checkup for %s...", env.Alias)

	config, err := OcmLogin(ocmToken, string(env.Url))
	if err != nil {
		return apisv1.ClusterCheckupResult{
			Success: "",
			Failure: fmt.Sprintf("failed to perform Clusters Checkup for the environment %s", env.Alias),
			Error:   fmt.Errorf("failed to login into %s with the provided ocm token: %w", env.Url, err),
		}
	}
	log.Default().Printf("Config. loaded. Environment: %s, Organization ID to be checked for: %s", env, env.OrgId)

	connection, err := ocm.NewConnection().Config(config).Build()
	if err != nil {
		return apisv1.ClusterCheckupResult{
			Success: "",
			Failure: fmt.Sprintf("failed to perform Clusters Checkup for the environment %s", env.Alias),
			Error:   fmt.Errorf("failed to establish a connection with OCM CLI: %w", err),
		}
	}
	defer connection.Close()

	staleClusterInstances, err := FetchStaleClusterInstances(connection, env.OrgId)
	result := apisv1.ClusterCheckupResult{
		Success: GenerateNotificationMessage(staleClusterInstances, env),
		Error:   nil,
	}
	if err != nil {
		result.Failure = fmt.Sprintf("Some errors were encountered while fetching the stale cluster instances: %s", err.Error())
	}

	log.Default().Printf("Successfully performed the clusters checkup for any stale clusters in %s environment", env.Alias)
	return result
}

package scan

import (
	"fmt"
	"log"
	"os"
	"strings"

	apisv1 "github.com/mt-sre/mtsre-clusters-checker/apis/v1"
	"github.com/mt-sre/mtsre-clusters-checker/internal"
	"github.com/spf13/cobra"
)

type flagpole struct {
	EnvsAndOrgIds  []string
	SlackChannelID string
	Config         string
	AsBackground   bool
	OcmToken       string
	SlackToken     string
}

func NewCommand() *cobra.Command {
	flags := &flagpole{}
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scans for clusters older than 24hr.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags)
		},
	}
	cmd.Flags().StringArrayVar(
		&flags.EnvsAndOrgIds, "envs-and-org-ids",
		[]string{},
		"key:value pairs denoting the environments and the corresponding organizations id for which you want to scan the clusters. Expected format <environment>:<organization-id>. Environment is one of ['integration', 'staging', 'production']. For example, `clusters-checker scan --envs-and-org-id staging:fooid1234 --envs-and-org-id production:barid1234` will scan the staging environment for the clusters owned by the organization `fooid1234` and production environment for the clusters owned by the organization `barid1234` older than 24hr.",
	)
	cmd.Flags().StringVar(
		&flags.SlackChannelID, "slack-channel-id",
		"",
		"ID of the slack channel where you want to receive the results of the scan.",
	)
	cmd.Flags().StringVar(
		&flags.OcmToken, "ocm-token",
		os.Getenv("OCM_TOKEN"),
		"OCM Token capable of querying OCM for the clusters belonging to your organization. Default value: OCM_TOKEN environment variable.",
	)
	cmd.Flags().StringVar(
		&flags.SlackToken, "slack-token",
		os.Getenv("SLACK_TOKEN"),
		"Slack Token capable of posting message on your target channel ID. Default value: SLACK_TOKEN environment variable.",
	)

	return cmd
}

func runE(flags *flagpole) error {
	ocmToken := flags.OcmToken
	if ocmToken == "" {
		return fmt.Errorf("OCM token not found to be provided")
	}

	var envs []apisv1.Environment
	for _, e := range flags.EnvsAndOrgIds {
		tokenizedEnv := strings.Split(e, ":")
		if len(tokenizedEnv) != 2 {
			return fmt.Errorf("%s found to be provided in unexpected format. Expected format is env:orgid", e)
		}

		var env apisv1.Environment
		switch tokenizedEnv[0] {
		case "integration":
			env = apisv1.Integration
		case "staging":
			env = apisv1.Staging
		case "production":
			env = apisv1.Production
		default:
			return fmt.Errorf("%s found to be provided with an unexpected env. Expected envs: ['integration', 'staging', 'production']", e)
		}
		env.OrgId = tokenizedEnv[1]

		envs = append(envs, env)
	}

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

	if err := internal.NotifyOnSlack(flags.SlackToken, flags.SlackChannelID, consolidatedMsg); err != nil {
		return fmt.Errorf("failed to notify about the stale cluster instances on slack: %w", err)
	}
	log.Println("Cluster Checkup performed successfully")
	return nil
}

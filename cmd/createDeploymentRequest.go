
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/fieldju/awesom-o/pkg/deployengine"
	"github.com/spf13/cobra"
	"log"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create-deployment-request",
	Short: "Creates a deployment request",
	Run: executeCreateDeploymentRequestCmd,
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func executeCreateDeploymentRequestCmd(cmd *cobra.Command, args []string) {
	request := &deployengine.K8sDeploymentRequest{
		Application: "demo-app",
		Account: "my-agent-configured-account",
		Namespace: "test",
		Manifests: []*deployengine.K8sManifest{
			{
				Name: "application-manifest",
				InlineValue: &deployengine.K8sManifestInlineValue{
					Value: "my-manifest here",
				},
			},
		},
		CanaryStrategy: &deployengine.K8sCanaryStrategy{
			Steps: []*deployengine.K8sCanaryStep{
				{
					SetWeightStep: &deployengine.SetWeightStep{
						Weight: 33,
					},
				},
				{
					PauseStep: &deployengine.PauseStep{
						Duration: 10,
						Unit: "minutes",
					},
				},
				{
					SetWeightStep: &deployengine.SetWeightStep{
						Weight: 33,
					},
				},
				{
					PauseStep: &deployengine.PauseStep{
						UntilApproved: true,
					},
				},
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		log.Fatalln("failed to marshal deploy engine request err:" + err.Error())
	}

	fmt.Println(string(jsonBytes))
}
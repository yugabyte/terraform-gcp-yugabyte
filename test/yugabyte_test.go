package yugabyte

import (
	"crypto/tls"
	"strings"
	"testing"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

func FindStringInResponse(statusCode int, body string) bool {
	if statusCode != 200 {
		return false
	}

	return strings.Contains(body, "YugaByte DB")
}

func TestYugaByteGcpTerraform(t *testing.T) {
	t.Parallel()

	yugabyteDir := test_structure.CopyTerraformFolderToTemp(t, "..", "../terraform-gcp-yugabyte")

	projectID := "<your-gcp-project-id>"
	credentials := "<your-gcp-credentials>"
	ssh_private_key := "<your-private-key>"
	ssh_public_key := "<you-public-key>"
	ssh_user := "<your-ssh-user>"
	region_name := "us-west1"
	cluster_name := "demo"
	node_count := 3
	yb_version := "2.0.0.0"

	terraformOptions := &terraform.Options{
		TerraformDir: yugabyteDir,

		Vars: map[string]interface{}{
			"project_id":      projectID,
			"credentials":     credentials,
			"region_name":     region_name,
			"ssh_private_key": ssh_private_key,
			"ssh_public_key":  ssh_public_key,
			"ssh_user":        ssh_user,
			"cluster_name":    cluster_name,
			"node_count":      node_count,
			"yb_version":      yb_version,
		},
	}

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndPlan(t, terraformOptions)
	terraform.Apply(t, terraformOptions)

	YugaByteURL := terraform.Output(t, terraformOptions, "ui")

	TlsConfig := tls.Config{}

	maxRetries := 30
	timeBetweenRetries := 5 * time.Second

	http_helper.HttpGetWithRetryWithCustomValidation(t, YugaByteURL, &TlsConfig, maxRetries, timeBetweenRetries, FindStringInResponse)

}

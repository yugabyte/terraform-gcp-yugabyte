package yugabyte

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/gruntwork-io/terratest/modules/gcp"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

var projectID, sshUser, yugabyteDir, publicIP, RemoteDir string

func FindStringInResponse(statusCode int, body string) bool {
	if statusCode != 200 {
		return false
	}
	return strings.Contains(body, "YugabyteDB")
}

func TestYugaByteGcpTerraform(t *testing.T) {
	t.Parallel()

	yugabyteDir := test_structure.CopyTerraformFolderToTemp(t, "..", "../terraform-gcp-yugabyte")
	keyPair := ssh.GenerateRSAKeyPair(t, 2048)
	maxRetries := 30
	timeBetweenRetries := 5 * time.Second

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, yugabyteDir)
		terraform.Destroy(t, terraformOptions)
		os.RemoveAll(yugabyteDir)
	})

	test_structure.RunTestStage(t, "SetUp", func() {
		terraformOptions := configureTerraformOptions(t, yugabyteDir)
		test_structure.SaveTerraformOptions(t, yugabyteDir, terraformOptions)
		terraform.InitAndApply(t, terraformOptions)
	})

	test_structure.RunTestStage(t, "validate", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, yugabyteDir)
		hosts := terraform.Output(t, terraformOptions, "hosts")
		YugaByteHosts := strings.Fields(strings.Trim(hosts, "[]\"\""))

		testYugaByteMasterURL(t, terraformOptions, maxRetries, timeBetweenRetries)
		testYugaByteTserverURL(t, terraformOptions, maxRetries, timeBetweenRetries)


		for _, host := range YugaByteHosts {

			host = strings.Trim(host, "\"\",")
			logger.Logf(t, "Host is :- %s", host)
			Intsnace, err := gcp.FetchInstanceE(t, projectID, host)
			logger.Logf(t, "Error While fetching instance :- %s", err)
			publicIP := Intsnace.GetPublicIp(t)
			Intsnace.AddSshKey(t, sshUser, keyPair.PublicKey)
			testYugaByteSSH(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP)
			testYugaByteYSQLSH(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP)
			testYugaByteCQLSH(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP)
			testYugaByteConf(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP, yugabyteDir)
			testYugaByteProcess(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP, "yb-master")
			testYugaByteProcess(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP, "yb-tserver")
			testYugaByteLogFile(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP, "master.out", "master")
			testYugaByteLogFile(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP, "master.err", "master")
			testYugaByteLogFile(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP, "tserver.out", "tserver")
			testYugaByteLogFile(t, terraformOptions, maxRetries, timeBetweenRetries, keyPair, publicIP, "tserver.err", "tserver")
			
		}

	})
}


func configureTerraformOptions(t *testing.T, yugabyteDir string) *terraform.Options {

	projectID = os.Getenv("GOOGLE_PROJECT_ID")
	credentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	sshUser = strings.ToLower(randomdata.FirstName(randomdata.Male))
	regionName := gcp.GetRandomRegion(t, projectID, nil, nil)
	clusterName := strings.ToLower(randomdata.FirstName(randomdata.Male))

	terraformOptions := &terraform.Options{
		TerraformDir: yugabyteDir,

		Vars: map[string]interface{}{
			"project_id":   projectID,
			"credentials":  credentials,
			"region_name":  regionName,
			"ssh_user":     sshUser,
			"cluster_name": clusterName,			
		},
	}

	ybVersion, ok := os.LookupEnv("TAG_NAME")
	if ok {
		terraformOptions.Vars["yb_version"]= ybVersion
	}
	return terraformOptions
}


func testYugaByteMasterURL(t *testing.T, terraformOptions *terraform.Options, maxRetries int, timeBetweenRetries time.Duration) {
	YugaByteURL := terraform.Output(t, terraformOptions, "master-ui")
	TLSConfig := tls.Config{}
	http_helper.HttpGetWithRetryWithCustomValidation(t, YugaByteURL, &TLSConfig, maxRetries, timeBetweenRetries, FindStringInResponse)
}

func testYugaByteTserverURL(t *testing.T, terraformOptions *terraform.Options, maxRetries int, timeBetweenRetries time.Duration) {
	YugaByteURL := terraform.Output(t, terraformOptions, "tserver-ui")

	TLSConfig := tls.Config{}
	http_helper.HttpGetWithRetryWithCustomValidation(t, YugaByteURL, &TLSConfig, maxRetries, timeBetweenRetries, FindStringInResponse)
}


func testYugaByteSSH(t *testing.T, terraformOptions *terraform.Options, maxRetries int, timeBetweenRetries time.Duration, keyPair *ssh.KeyPair, publicIP string) {

	sshHost := ssh.Host{
		Hostname:    string(publicIP),
		SshKeyPair:  keyPair,
		SshUserName: sshUser,
	}
	sampleText := "Hello World"
	retry.DoWithRetry(t, "Attempting to SSH", maxRetries, timeBetweenRetries, func() (string, error) {
		output, err := ssh.CheckSshCommandE(t, sshHost, fmt.Sprintf("echo '%s'", sampleText))
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(sampleText) != strings.TrimSpace(output) {
			return "", fmt.Errorf("Expected: %s. Got: %s\n", sampleText, output)
		}
		return "", nil
	})

}

func testYugaByteYSQLSH(t *testing.T, terraformOptions *terraform.Options, maxRetries int, timeBetweenRetries time.Duration, keyPair *ssh.KeyPair, publicIP string) {

	commandConnectYSQLSH := "cd " + filepath.Join("/home/", sshUser, "/yugabyte-db/tserver") + " && ./bin/ysqlsh  --echo-queries -h " + string(publicIP)
	sshHost := ssh.Host{
		Hostname:    string(publicIP),
		SshKeyPair:  keyPair,
		SshUserName: sshUser,
	}

	retry.DoWithRetry(t, "Conecting YSQLSH", maxRetries, timeBetweenRetries, func() (string, error) {
		output, err := ssh.CheckSshCommandE(t, sshHost, commandConnectYSQLSH)
		if err != nil {
			return "", err
		}
		logger.Logf(t, "Output of Ysql command :- %s", output)
		return output, nil
	})

}

func testYugaByteCQLSH(t *testing.T, terraformOptions *terraform.Options, maxRetries int, timeBetweenRetries time.Duration, keyPair *ssh.KeyPair, publicIP string) {

	commandConnectCQLSH := "cd " + filepath.Join("/home/", sshUser, "/yugabyte-db/tserver") + " && ./bin/cqlsh " + string(publicIP) + " 9042"
	sshHost := ssh.Host{
		Hostname:    string(publicIP),
		SshKeyPair:  keyPair,
		SshUserName: sshUser,
	}

	retry.DoWithRetry(t, "Conecting CQLSH", maxRetries, timeBetweenRetries, func() (string, error) {
		output, err := ssh.CheckSshCommandE(t, sshHost, commandConnectCQLSH)
		if err != nil {
			return "", err
		}
		logger.Logf(t, "Output of Cql command :- %s", output)
		return output, nil
	})

}

func testYugaByteConf(t *testing.T, terraformOptions *terraform.Options, maxRetries int, timeBetweenRetries time.Duration, keyPair *ssh.KeyPair, publicIP string, yugabyteDir string) {
	RemoteDir := filepath.Join("/home/", sshUser)
	sshHost := ssh.Host{
		Hostname:    string(publicIP),
		SshKeyPair:  keyPair,
		SshUserName: sshUser,
	}

	getTserverConf := ssh.ScpDownloadOptions{
		RemoteHost: sshHost,
		RemoteDir:  filepath.Join(RemoteDir, "/yugabyte-db/tserver/conf/"),
		LocalDir:   yugabyteDir,
	}

	getFileError := ssh.ScpDirFromE(t, getTserverConf, false)

	if getFileError != nil {
		logger.Logf(t, "We got error while getting file from server :- %s", getFileError)
	}

	assert.FileExists(t, filepath.Join(yugabyteDir, "/server.conf"))
}

func testYugaByteProcess(t *testing.T, terraformOptions *terraform.Options, maxRetries int, timeBetweenRetries time.Duration,  keyPair *ssh.KeyPair, publicIP string, processName string) {
	sshHost := ssh.Host{
		Hostname:    string(publicIP),
		SshKeyPair:  keyPair,
		SshUserName: sshUser,
	}
	retry.DoWithRetry(t, "Checking process "+processName, maxRetries, timeBetweenRetries, func() (string, error) {
		output, err := ssh.CheckSshCommandE(t, sshHost, fmt.Sprintf("pgrep --list-name '%s'", processName))
		if err != nil {
			return "", err
		}
		if ! strings.Contains(output, processName) {
			return "", fmt.Errorf("Expected: %s. Got: %s\n", processName, output)
		}
		return "", nil
	})

}

func testYugaByteLogFile(t *testing.T, terraformOptions *terraform.Options, maxRetries int, timeBetweenRetries time.Duration,  keyPair *ssh.KeyPair, publicIP string, logFile string, logDir string) {
	RemoteDir:= filepath.Join("/home", sshUser, "yugabyte-db", logDir)
	sshHost := ssh.Host{
		Hostname:    string(publicIP),
		SshKeyPair:  keyPair,
		SshUserName: sshUser,
	}
	
	retry.DoWithRetry(t, "Checking log file "+logFile, maxRetries, timeBetweenRetries, func() (string, error) {
		output, err := ssh.CheckSshCommandE(t, sshHost, fmt.Sprintf("ls '%s' | grep '%s'",RemoteDir, logFile))
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(logFile) != strings.TrimSpace(output) {
			return "", fmt.Errorf("Expected: %s. Got: %s\n", logFile, output)
		}
		return "", nil
	})

}
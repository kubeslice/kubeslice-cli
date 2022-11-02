package internal

import (
	"fmt"
	"log"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

const imagePullSecretsTemplate = `

imagePullSecrets:
  repository: %s
  username: %s
  password: %s
  %s

`

func AddHelmCharts(ApplicationConfiguration *ConfigurationSpecs) {
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	// helm repo add avesha https://kubeslice.github.io/kubeslice/
	util.Printf("\nAdding KubeSlice Helm Charts...")

	addHelmChart(ApplicationConfiguration)
	util.Printf("%s Successfully added helm repo %s : %s", util.Tick, hc.RepoAlias, hc.RepoUrl)
	time.Sleep(200 * time.Millisecond)

	updateHelmChart()
	util.Printf("%s Successfully updated helm repo", util.Tick)
	time.Sleep(200 * time.Millisecond)

	util.Printf("%s Successfully added helm charts.\n", util.Tick)
}

func addHelmChart(ApplicationConfiguration *ConfigurationSpecs) {
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	repoAddCommands := make([]string, 0)
	repoAddCommands = append(repoAddCommands, "repo", "add", hc.RepoAlias, hc.RepoUrl, "--force-update")
	if hc.HelmUsername != "" && hc.HelmPassword != "" {
		repoAddCommands = append(repoAddCommands, "--pass-credentials", "--username", hc.HelmUsername, "--password", hc.HelmPassword)
	}
	err := util.RunCommand("helm", repoAddCommands...)
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func updateHelmChart() {
	err := util.RunCommand("helm", "repo", "update")
	if err != nil {
		log.Fatalf("Process failed %v", err)
	}
}

func generateImagePullSecretsValue(ImagePullSecret ImagePullSecrets) string {
	imagePullSecretsValue := ""
	ips := ImagePullSecret
	if ips.Registry != "" && ips.Username != "" && ips.Password != "" {
		email := ""
		if ips.Email != "" {
			email = "email: " + ips.Email
		}
		imagePullSecretsValue = fmt.Sprintf(imagePullSecretsTemplate, ips.Registry, ips.Username, ips.Password, email)
	}
	return imagePullSecretsValue
}

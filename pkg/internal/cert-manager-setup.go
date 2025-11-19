package internal

import (
	"fmt"
	"time"

	"github.com/kubeslice/kubeslice-cli/util"
)

// Function variable for testing - Allows mocking PodVerification in tests
var podVerificationFunc = PodVerification

func InstallCertManager(ApplicationConfiguration *ConfigurationSpecs) {

	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	util.Printf("\nInstall Cert Manager to Controller Cluster...")

	installCertManager(cc.ControllerCluster, hc)
	util.Printf("%s Successfully installed helm chart %s/%s", util.Tick, hc.RepoAlias, hc.CertManagerChart.ChartName)
	util.Sleep(200 * time.Millisecond)

	util.Printf("%s Waiting for Cert Manager Pods to be Healthy...", util.Wait)
	podVerificationFunc("Waiting for Cert Manager Pods to be Healthy", cc.ControllerCluster, "cert-manager")

	util.Printf("%s Successfully installed cert manager.\n", util.Tick)

}
func UninstallCertManager(ApplicationConfiguration *ConfigurationSpecs) {

	cc := ApplicationConfiguration.Configuration.ClusterConfiguration
	hc := ApplicationConfiguration.Configuration.HelmChartConfiguration
	util.Printf("\nUninstalling Cert Manager...")

	err := uninstallCertManager(cc.ControllerCluster, hc)
	if err == nil {
		util.Printf("%s Successfully uninstalled cert manager.\n", util.Tick)
	} else {
		util.Printf("%s Failed to uninstall cert manager.\n", util.Cross)
	}

}

func installCertManager(cluster Cluster, hc HelmChartConfiguration) {
	args := make([]string, 0)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "upgrade", "-i", "cert-manager", fmt.Sprintf("%s/%s", hc.RepoAlias, hc.CertManagerChart.ChartName), "--namespace", "cert-manager", "--create-namespace", "--set", "installCRDs=true")
	if hc.CertManagerChart.Version != "" {
		args = append(args, "--version", hc.CertManagerChart.Version)
	}
	err := util.RunCommand("helm", args...)
	if err != nil {
		util.Fatalf("Process failed %v", err)
	}
}
func uninstallCertManager(cluster Cluster, hc HelmChartConfiguration) error {
	args := make([]string, 0)
	args = append(args, "--kube-context", cluster.ContextName, "--kubeconfig", cluster.KubeConfigPath, "uninstall", "cert-manager", "--namespace", "cert-manager")

	err := util.RunCommand("helm", args...)
	return err

}

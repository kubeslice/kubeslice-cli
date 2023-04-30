## kubeslice-cli install

Installs workloads to run KubeSlice

### Synopsis

Installs the required workloads to run KubeSlice Controller and KubeSlice Worker.
	Additional example applications are also installed in demo profiles to showcase the
	KubeSlice functionality

```
kubeslice-cli install [flags]
```

### Options

```
  -h, --help                help for install
  -p, --profile string      <profile-value>
                            The profile for installation/uninstallation.
                            Supported values:
                            	- full-demo:
                            		Showcases the KubeSlice inter-cluster connectivity by spawning
                            		3 Kind Clusters, including 1 KubeSlice Controller and 2 KubeSlice Workers, 
                            		and installing iPerf application to generate network traffic.
                            	- minimal-demo:
                            		Sets up 3 Kind Clusters, including 1 KubeSlice Controller and 2 KubeSlice Workers. 
                            		Generates the KubernetesManifests for user to manually apply, and verify 
                            		the functionality
                            	- enterprise-demo:
                            		Showcases the KubeSlice Enterprise functionality by spawning
                            		3 Kind Clusters, including 1 KubeSlice Controller and 2 KubeSlice Workers, 
                            		installing the enterprise charts for Controller and Worker with KubeSlice Manager (UI),
                            		and installing iPerf application to generate network traffic. 
                            		Ensure that the imagePullSecrets (username and password) are set as environment variables.
                            
                            		KUBESLICE_IMAGE_PULL_USERNAME : optional : Default 'aveshaenterprise'
                            		KUBESLICE_IMAGE_PULL_PASSWORD : required
                            
                            Cannot be used with --config flag.
  -s, --skip strings        Skips the installation steps (comma-seperated). 
                            Supported values:
                            	- kind: Skips the creation of kind clusters
                            	- calico: Skips the installation of Calico
                            	- controller: Skips the installation of KubeSlice Controller
                            	- worker-registration: Skips the registration of KubeSlice Workers on the Controller
                            	- worker: Skips the installation of KubeSlice Worker
                            	- demo: Skips the installation of additional example applications
                            	- ui: Skips the installtion of enterprise UI components (Kubeslice-Manager)
                            	- prometheus: Skips the installation of prometheus
      --with-cert-manager   Installs Cert-Manager for kubeslice controller (for versions < 0.7.0)
```

### Options inherited from parent commands

```
  -c, --config string   <path-to-topology-configuration-yaml-file>
                        	The yaml file with topology configuration. 
                        	Refer: https://github.com/kubeslice/kubeslice-cli/blob/master/samples/template.yaml
```

### SEE ALSO

* [kubeslice-cli](kubeslice-cli.md)	 - kubeslice-cli - a simple CLI for KubeSlice Operations
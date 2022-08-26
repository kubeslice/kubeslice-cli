## kubeslice-cli

kubeslice-cli - A simple CLI for KubeSlice Operations

### Synopsis

kubeslice-cli - A simple CLI for KubeSlice Operations
    
Use kubeslice-cli to install/uninstall required workloads to run KubeSlice Controller and KubeSlice Worker.
Additional example applications can also be installed in demo profiles to showcase the
KubeSlice functionality

```
kubeslice-cli [flags]
```

### Options

```
  -c, --config string   <path-to-topology-configuration-yaml-file>
                        	The yaml file with topology configuration. 
                        	Refer: https://github.com/kubeslice/slicectl/blob/master/samples/template.yaml
  -h, --help            help for kubeslice-cli
  -v, --version         version for kubeslice-cli
```

### SEE ALSO

* [kubeslice-cli completion](kubeslice-cli_completion.md)	 - Generate the autocompletion script for the specified shell
* [kubeslice-cli create](kubeslice-cli_create.md)	 - Create Kubeslice resources.
* [kubeslice-cli delete](kubeslice-cli_delete.md)	 - Delete Kubeslice resources.
* [kubeslice-cli describe](kubeslice-cli_describe.md)	 - Describe Kubeslice resources.
* [kubeslice-cli edit](kubeslice-cli_edit.md)	 - Edit Kubeslice resources.
* [kubeslice-cli get](kubeslice-cli_get.md)	 - Get Kubeslice resources.
* [kubeslice-cli install](kubeslice-cli_install.md)	 - Installs workloads to run KubeSlice
* [kubeslice-cli uninstall](kubeslice-cli_uninstall.md)	 - Deletes the Kind Clusters used for the demo.



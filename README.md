# `kubeslice-cli` for simplified KubeSlice Operations

![GitHub release (latest by date)](https://img.shields.io/github/v/release/kubeslice/kubeslice-cli?style=for-the-badge)

This repository provides `kubeslice-cli` tool.
[Install &rarr;](#installation)

## Installation

`kubeslice-cli` is a Go Lang executable utility which helps perform the KubeSlice operations like installation, demo, etc. 
on Kubernetes Clusters. You can download a suitable version of the installer from the [**Releases page
&rarr;**](https://github.com/kubeslice/kubeslice-cli/releases)


## Synopsis

kubeslice-cli - A simple CLI for KubeSlice Operations
    
Use kubeslice-cli to install/uninstall required workloads to run KubeSlice Controller and KubeSlice Worker.
Additional example applications can also be installed in demo profiles to showcase the
KubeSlice functionality

### Usage
```
  kubeslice-cli [command] [flags]
```

### Available Commands
```
  create      Create Kubeslice resources.
  delete      Delete Kubeslice resources.
  describe    Describe Kubeslice resources.
  edit        Edit Kubeslice resources.
  get         Get Kubeslice resources.
  install     Installs workloads to run KubeSlice
  register    Register a Kubeslice worker cluster.
  show-health Show health status of KubeSlice resources.
  uninstall   Performs cleanup of Kubeslice components.
  help        Help about any command

```

### Options

```
  -c, --config string   <path-to-topology-configuration-yaml-file>
                        	The yaml file with topology configuration. 
                        	Refer: https://github.com/kubeslice/kubeslice-cli/blob/master/samples/template.yaml
  -h, --help            help for kubeslice-cli
  -v, --version         version for kubeslice-cli
```

### SEE ALSO

* [kubeslice-cli create](doc/kubeslice-cli_create.md)	 - Create Kubeslice resources.
* [kubeslice-cli delete](doc/kubeslice-cli_delete.md)	 - Delete Kubeslice resources.
* [kubeslice-cli describe](doc/kubeslice-cli_describe.md)	 - Describe Kubeslice resources.
* [kubeslice-cli edit](doc/kubeslice-cli_edit.md)	 - Edit Kubeslice resources.
* [kubeslice-cli get](doc/kubeslice-cli_get.md)	 - Get Kubeslice resources.
* [kubeslice-cli install](doc/kubeslice-cli_install.md)	 - Installs workloads to run KubeSlice.
* [kubeslice-cli register](doc/kubeslice-cli_register.md)	 - Register a Kubeslice worker cluster.
* [kubeslice-cli show-health](doc/kubeslice-cli_show-health.md)	 - Show health status of KubeSlice resources.
* [kubeslice-cli uninstall](doc/kubeslice-cli_uninstall.md)	 - Performs cleanup of Kubeslice components.



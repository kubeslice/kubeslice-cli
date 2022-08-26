# `kubeslice-cli` for simplified KubeSlice Operations

[comment]: <> (![Latest GitHub release]&#40;https://img.shields.io/github/release/kubeslice/slicectl.svg&#41;)

This repository provides `kubeslice-cli` tool.
[Install &rarr;](#installation)

## Installation

`kubeslice-cli` is a Go Lang executable utility which helps perform the KubeSlice operations like installation, demo, etc. 
on Kubernetes Clusters. You can download a suitable version of the installer from the [**Releases page
&rarr;**](https://github.com/kubeslice/slicectl/releases)


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
  uninstall   Deletes the Kind Clusters used for the demo.
  help        Help about any command

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

* [kubeslice-cli create](doc/kubeslice-cli_create.md)	 - Create Kubeslice resources.
* [kubeslice-cli delete](doc/kubeslice-cli_delete.md)	 - Delete Kubeslice resources.
* [kubeslice-cli describe](doc/kubeslice-cli_describe.md)	 - Describe Kubeslice resources.
* [kubeslice-cli edit](doc/kubeslice-cli_edit.md)	 - Edit Kubeslice resources.
* [kubeslice-cli get](doc/kubeslice-cli_get.md)	 - Get Kubeslice resources.
* [kubeslice-cli install](doc/kubeslice-cli_install.md)	 - Installs workloads to run KubeSlice
* [kubeslice-cli uninstall](doc/kubeslice-cli_uninstall.md)	 - Deletes the Kind Clusters used for the demo.



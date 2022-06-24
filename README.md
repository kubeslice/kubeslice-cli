# `slicectl` for simplified KubeSlice Operations

[comment]: <> (![Latest GitHub release]&#40;https://img.shields.io/github/release/kubeslice/slicectl.svg&#41;)

This repository provides `slicectl` tool.
[Install &rarr;](#installation)

## Installation

`slicectl` is a Go Lang executable utility which helps perform the KubeSlice operations like installation, demo, etc. 
on Kubernetes Clusters. You can download a suitable version of the installer from the [**Releases page
&rarr;**](https://github.com/kubeslice/slicectl/releases)

## Usage

slicectl for KubeSlice Operations

Usage

```
slicectl <options> <command>
```

Options:

```
  --profile=<profile-value>
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

  --config=<path-to-topology-configuration-yaml>
      The yaml file with topology configuration.
      Refer: https://github.com/kubeslice/slicectl/blob/master/samples/template.yaml
```

Commands:
```
  install
      Creates 3 Kind Clusters, sets-up KubeSlice Controller, KubeSlice Worker,
      and iperf example application.
      Once the setup is done, prints the instructions on how to create a slice
      and verify the connectivity.

  uninstall
      Deletes the Kind Clusters used for the demo.

  help
      Prints this help menu.
```

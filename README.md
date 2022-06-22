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
slicectl [OPTIONS]
```

Options:

```
--help
        Prints the help menu

--full-install
        Creates 3 Kind Clusters, sets-up KubeSlice Controller, 
        KubeSlice Worker,a demo slice, and iperf example application

--minimal-install
        Creates 3 Kind Clusters, sets-up KubeSlice Controller, KubeSlice Worker,
        and iperf example application.
        Once the setup is done, prints the instructions on how to create a slice
        and verify the connectivity.

--uninstall
        Deletes the 3 Kind Clusters, but retains the kubeslice configuration directory.

--cleanup
        Deletes the 3 Kind Clusters and kubeslice directory.
```

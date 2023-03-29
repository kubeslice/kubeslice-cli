## kubeslice-cli uninstall

Performs cleanup of Kubeslice components.

```
kubeslice-cli uninstall [flags]
```

### Options

```
  -a, --all            Uninstalls all components (Worker, Controller, UI)
      --cert-manager   Uninstalls Cert Manager (required for controller version < 0.7.0)
  -h, --help           help for uninstall
  -u, --ui             Uninstalls enterprise UI components (Kubeslice-Manager)
```

### Options inherited from parent commands

```
  -c, --config string   <path-to-topology-configuration-yaml-file>
                        	The yaml file with topology configuration. 
                        	Refer: https://github.com/kubeslice/kubeslice-cli/blob/master/samples/template.yaml
```

### SEE ALSO

* [kubeslice-cli](kubeslice-cli.md)	 - kubeslice-cli - a simple CLI for KubeSlice Operations



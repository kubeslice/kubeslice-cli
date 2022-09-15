## kubeslice-cli edit

Edit Kubeslice resources.

### Synopsis

The edit command allows you to directly edit any Kubeslice resource you can retrieve via the command line tools. It will open the editor defined by your KUBE_EDITOR, or EDITOR environment variables, or fall back to ‘vi’ for Linux or ‘notepad’ for Windows. You can edit multiple objects, although changes are applied one at a time. The command accepts filenames as well as command line arguments, although the files you point to must be previously saved versions of resources.
	The default format is YAML.
	In the event an error occurs while updating, a temporary file will be created on disk that contains your unapplied changes. The most common error when updating a resource is another editor changing the resource on the server. When this occurs, you will have to apply your changes to the newer version of the resource, or update your temporary saved copy to include the latest resource version.

```
kubeslice-cli edit [flags]
```

### Options

```
  -f, --filename string    Filename, directory, or URL to file to use to create the resource
  -h, --help               help for edit
  -n, --namespace string   namespace
```

### Options inherited from parent commands

```
  -c, --config string   <path-to-topology-configuration-yaml-file>
                        	The yaml file with topology configuration. 
                        	Refer: https://github.com/kubeslice/slicectl/blob/master/samples/template.yaml
```

### SEE ALSO

* [kubeslice-cli](kubeslice-cli.md)	 - kubeslice-cli - a simple CLI for KubeSlice Operations



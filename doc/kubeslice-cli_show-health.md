# kubeslice-cli show-health

Show health status of KubeSlice resources.

## Synopsis

The `show-health` command allows you to view the health status of KubeSlice slices. It provides detailed information about the health and status of slice configurations.

```bash
kubeslice-cli show-health [resource-type] [resource-name] [flags]
```

## Examples

### Show health for a specific slice
```bash
kubeslice-cli show-health slice my-slice -n kubeslice-demo
```

### Show health for all slices
```bash
kubeslice-cli show-health slice -A -n kubeslice-demo
```

### Show health with JSON output
```bash
kubeslice-cli show-health slice my-slice -n kubeslice-demo -o json
```

### Show health with YAML output
```bash
kubeslice-cli show-health slice my-slice -n kubeslice-demo -o yaml
```

## Available Resource Types

- `slice` - Show health status of KubeSlice slice configurations

## Flags

- `-A, --all` - Show health for all slices (when used with slice resource type)
- `-n, --namespace string` - Namespace where the slice is located (required)
- `-o, --output string` - Output format (json, yaml)

## Global Flags

- `-c, --config string` - Path to topology configuration YAML file

## Health Status Information

The command displays the following health information:

- **Status**: Ready/Not Ready/Unknown
- **Conditions**: Available health conditions
- **Slice Subnet**: Configuration status
- **Clusters**: Connection status
- **Raw Output**: Complete JSON/YAML output from kubectl

## Usage Notes

1. The `--namespace` flag is required for all show-health operations
2. When using `--all` flag, you don't need to specify a resource name
3. The command uses kubectl to fetch slice configuration data
4. Health status is determined by parsing the slice configuration status
5. Raw output is always displayed for debugging purposes

## Error Handling

- If the slice doesn't exist, an error will be displayed
- If the namespace doesn't exist, an error will be displayed
- If kubectl is not available, an error will be displayed
- Network connectivity issues will be reported

## Related Commands

- `kubeslice-cli get slice` - Get slice configuration details
- `kubeslice-cli describe slice` - Describe slice configuration
- `kubeslice-cli create slice` - Create a new slice
- `kubeslice-cli delete slice` - Delete a slice 
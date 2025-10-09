package internal

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

func MergeMaps(dest, src map[interface{}]interface{}) map[interface{}]interface{} {
	// If dest is nil, create a new map
	if dest == nil {
		dest = make(map[interface{}]interface{})
	}

	for k, v := range src {
		if d, ok := dest[k]; ok {
			switch d.(type) {
			case map[interface{}]interface{}:
				dest[k] = MergeMaps(d.(map[interface{}]interface{}), v.(map[interface{}]interface{}))
			default:
				dest[k] = v
			}
		} else {
			dest[k] = v
		}
	}
	return dest
}

func GenerateValuesFile(filePath string, hc *HelmChart, defaults string) error {
	if hc == nil {
		return fmt.Errorf("helm chart cannot be nil")
	}

	valuesMap := make(map[interface{}]interface{})
	for k, v := range hc.Values {
		keys := strings.Split(k, ".")
		currentMap := valuesMap
		for i, key := range keys {
			if i == len(keys)-1 {
				currentMap[key] = v
			} else {
				if currentMap[key] == nil {
					currentMap[key] = make(map[interface{}]interface{})
				}
				currentMap = currentMap[key].(map[interface{}]interface{})
			}
		}
	}

	defaultsMap := make(map[interface{}]interface{})
	if err := yaml.Unmarshal([]byte(defaults), &defaultsMap); err != nil {
		return fmt.Errorf("error parsing defaults: %v", err)
	}

	mergedMap := MergeMaps(defaultsMap, valuesMap)

	finalData, err := yaml.Marshal(mergedMap)
	if err != nil {
		return fmt.Errorf("error encoding final data as YAML: %v", err)
	}

	if err := ioutil.WriteFile(filePath, finalData, 0644); err != nil {
		return fmt.Errorf("error writing values file: %v", err)
	}

	return nil
}

package internal

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

func mergeMaps(dest, src map[interface{}]interface{}) map[interface{}]interface{} {
	for k, v := range src {
		if d, ok := dest[k]; ok {
			switch d.(type) {
			case map[interface{}]interface{}:
				dest[k] = mergeMaps(d.(map[interface{}]interface{}), v.(map[interface{}]interface{}))
			default:
				dest[k] = v
			}
		} else {
			dest[k] = v
		}
	}
	return dest
}

func generateValuesFile(filePath string, hc *HelmChart, defaults string) error {
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

	mergedMap := mergeMaps(valuesMap, defaultsMap)

	finalData, err := yaml.Marshal(mergedMap)
	if err != nil {
		return fmt.Errorf("error encoding final data as YAML: %v", err)
	}

	if err := ioutil.WriteFile(filePath, finalData, 0644); err != nil {
		return fmt.Errorf("error writing values file: %v", err)
	}

	return nil
}

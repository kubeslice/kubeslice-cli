package util

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	Cross = string(rune(0x274c))
	Tick  = string(rune(0x2714))
	Check = string(rune(0x2714))
	Wait  = string(rune(0x267B))
	Run   = string(rune(0x1F3C3))
	Warn  = string(rune(0x26A0))
	Lock  = string(rune(0x1F512))
	Globe = string(rune(0x1F310))
	Info  = string(rune(0x2139))
)

func Printf(format string, a ...interface{}) {
	if len(a) > 0 {
		fmt.Printf(format+"\n", a...)
	} else {
		fmt.Println(format)
	}
}

func Fatalf(format string, a ...interface{}) {
	if len(a) > 0 {
		fmt.Printf(format+"\n", a...)
	} else {
		fmt.Println(format + "\n")
	}
	os.Exit(1)
}

func PrintJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		Printf("Error marshaling to JSON: %v", err)
		return
	}
	fmt.Println(string(jsonData))
}

func PrintYAML(data interface{}) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		Printf("Error marshaling to YAML: %v", err)
		return
	}
	fmt.Println(string(yamlData))
}

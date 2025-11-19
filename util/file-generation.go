package util

import (
	"errors"
	"os"
)

func CreateDirectoryPath(path string) {
	if _, err := FileSystem.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := FileSystem.MkdirAll(path, os.ModePerm)
		if err != nil {
			Fatalf("%s Failed to create kubeslice directory to generate configuration files.", Cross)
		}
	}
}

func DumpFile(template, filename string) {
	data := []byte(template)
	err := FileSystem.WriteFile(filename, data, 0644)
	if err != nil {
		Fatalf("%s Failed to write %s: %v", Cross, filename, err)
	}
}

package util

import (
	"errors"
	"os"
)

func CreateDirectoryPath(path string) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			Fatalf("%s Failed to create kubeslice directory to generate configuration files.", Cross)
		}
	}
}

func DumpFile(template, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		Fatalf("%s Failed to create %s", Cross, filename)
	}
	defer f.Close()
	data := []byte(template)
	_, err2 := f.Write(data)
	if err2 != nil {
		Fatalf("%s Failed to write %s", Cross, filename)
	}
}
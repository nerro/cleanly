package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	projectFile   = ".project"
	classpathFile = ".classpath"
	settingsDir   = ".settings"
)

func main() {
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	var eclipseFiles []string

	filepath.Walk(workingDir, func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir() && fileInfo.Name() == settingsDir {
			eclipseFiles = append(eclipseFiles, path)
		}
		if !fileInfo.IsDir() && fileInfo.Name() == projectFile {
			eclipseFiles = append(eclipseFiles, path)
		}
		if !fileInfo.IsDir() && fileInfo.Name() == classpathFile {
			eclipseFiles = append(eclipseFiles, path)
		}
		return nil
	})

	for _, file := range eclipseFiles {
		err := os.RemoveAll(file)
		if err != nil {
			fmt.Println(err)
		}
	}
}

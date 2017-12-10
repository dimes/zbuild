// Package copyutil contains code for copying files from one location to another
package copyutil

import (
	"fmt"
	"github.com/dimes/zbuild/buildlog"
	"io"
	"os"
	"path/filepath"
)

// Copy copies the source to the destination. If source is a directory, it will be copied
// recursively.
func Copy(source, destination string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("Error getting file info for source %s: %+v", source, err)
	}

	if !info.IsDir() {
		return copyFile(source, info, destination)
	}

	return copyDirectory(source, destination)
}

func copyDirectory(source, destination string) error {
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("Error determining relative path to %s: %+v", path, err)
		}

		outputPath := filepath.Join(destination, relativePath)
		if info.IsDir() {
			buildlog.Debugf("Creating output directory %s", outputPath)
			if err := os.MkdirAll(outputPath, info.Mode()); err != nil {
				return fmt.Errorf("Error creating directory %s: %+v", outputPath, err)
			}
			return nil
		}

		return copyFile(path, info, outputPath)
	})
}

func copyFile(source string, info os.FileInfo, destination string) error {
	buildlog.Debugf("Copying file %s to %s", source, destination)
	in, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("Error opening %s: %+v", source, err)
	}
	defer in.Close()

	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("Error opening %s: %+v", destination, err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("Error copying %s to %s: %+v", source, destination, err)
	}

	return nil
}

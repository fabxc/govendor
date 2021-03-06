// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kardianos/govendor/internal/pathos"
)

// CopyPackage copies the files from the srcPath to the destPath, destPath
// folder and parents are are created if they don't already exist.
func CopyPackage(destPath, srcPath string, ignoreFiles []string) error {
	if pathos.FileStringEquals(destPath, srcPath) {
		return fmt.Errorf("Attempting to copy package to same location %q.", destPath)
	}
	err := os.MkdirAll(destPath, 0777)
	if err != nil {
		return err
	}

	// Ensure the dest is empty of files.
	destDir, err := os.Open(destPath)
	if err != nil {
		return err
	}

	fl, err := destDir.Readdir(-1)
	destDir.Close()
	if err != nil {
		return err
	}
	for _, fi := range fl {
		if fi.IsDir() {
			continue
		}
		err = os.Remove(filepath.Join(destPath, fi.Name()))
		if err != nil {
			return err
		}
	}

	// Copy files into dest.
	srcDir, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	fl, err = srcDir.Readdir(-1)
	srcDir.Close()
	if err != nil {
		return err
	}
fileLoop:
	for _, fi := range fl {
		if fi.IsDir() {
			continue
		}
		if fi.Name()[0] == '.' {
			continue
		}
		for _, ignore := range ignoreFiles {
			if pathos.FileStringEquals(fi.Name(), ignore) {
				continue fileLoop
			}
		}
		err = copyFile(
			filepath.Join(destPath, fi.Name()),
			filepath.Join(srcPath, fi.Name()),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(destPath, srcPath string) error {
	ss, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(dest, src)
	// Close before setting mod and time.
	dest.Close()
	if err != nil {
		return err
	}
	err = os.Chmod(destPath, ss.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(destPath, ss.ModTime(), ss.ModTime())
}

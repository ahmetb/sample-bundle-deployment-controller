// Copyright 2022 Ahmet Alp Balkan

package release

import (
	"io/fs"
	"path/filepath"
	"testing"
)

func TestChecksum(t *testing.T) {
	sum, err := ReleaseBundles["rapid"].Checksum()
	if err != nil {
		t.Fatal(err)
	}
	if sum == "" {
		t.Fatal("empty sum")
	}
	t.Log(sum)
}

func TestWriteTo(t *testing.T) {
	tempDir := t.TempDir()
	err := ReleaseBundles["rapid"].WriteTo(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	var files int
	err = filepath.WalkDir(tempDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files++
		t.Log(path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 {
		t.Fatal("no files were found in target dir")
	}
}

// Copyright 2022 Ahmet Alp Balkan

package release

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	//go:embed stable/*
	stableManifests embed.FS

	//go:embed rapid/*
	rapidManifests embed.FS

	ReleaseBundles = map[string]ReleaseBundle{
		"rapid": {
			bundle:                 rapidManifests,
			sourceDir:              "rapid",
			DeploymentResourceName: "nginx-deployment",
		},
		"stable": {
			bundle:                 stableManifests,
			sourceDir:              "stable",
			DeploymentResourceName: "nginx-deployment",
		},
	}
)

type ReleaseBundle struct {
	bundle                 fs.FS
	sourceDir              string
	DeploymentResourceName string
}

// Checksum gives a sha256sum of all files in filepath.WalkDir order.
func (r ReleaseBundle) Checksum() (string, error) {
	h := sha256.New()
	var files int
	err := fs.WalkDir(r.bundle, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files++
		f, err := r.bundle.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(h, f)
		return err
	})
	if files == 0 {
		return "", fmt.Errorf("no files were checksummed")
	}
	if err != nil {
		return "", nil
	}

	return fmt.Sprintf("%x", sha256.Sum256(h.Sum(nil))), nil
}

// WriteTo writes the release bundle contents to specified directory
func (r ReleaseBundle) WriteTo(dir string) error {
	srcBundle, err := fs.Sub(r.bundle, r.sourceDir)
	if err != nil {
		return err
	}

	return fs.WalkDir(srcBundle, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fp := filepath.Join(dir, path)
		if d.IsDir() {
			return os.MkdirAll(fp, 0755)
		}
		f, err := os.OpenFile(fp, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		inFile, err := srcBundle.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, inFile)
		return err
	})
}

// Copyright 2022 Ahmet Alp Balkan

package release

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

var (
	//go:embed bundles/*
	bundleManifests embed.FS

	ReleaseBundles = map[string]ReleaseBundle{}
)

type ReleaseBundle struct {
	bundle fs.FS
}

func init() {
	// load embedded release bundles
	var err error
	for _, name := range []string{"rapid", "stable"} {
		ReleaseBundles[name], err = extractBundle(bundleManifests, name)
		if err != nil {
			panic(err)
		}
	}
}

func extractBundle(embeddedBundles embed.FS, name string) (ReleaseBundle, error) {
	subFS, err := fs.Sub(embeddedBundles, path.Join("bundles", name))
	if err != nil {
		return ReleaseBundle{}, fmt.Errorf("failed to load embedded bundle %q: :%w", name, err)
	}
	return ReleaseBundle{bundle: subFS}, nil
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
	return fs.WalkDir(r.bundle, ".", func(path string, d fs.DirEntry, err error) error {
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

		inFile, err := r.bundle.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, inFile)
		return err
	})
}

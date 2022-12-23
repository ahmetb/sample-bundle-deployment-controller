// Copyright 2022 Ahmet Alp Balkan

package deployer

import (
	"bundle-deployment-controller/controller/release"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type ManifestBuilder struct {
	releaseBundle release.ReleaseBundle
	kustomization kustomization
}

type ManifestBuilderOption func(*ManifestBuilder)

type kustomization struct {
	manifestPaths []string
	patches       []string
	namePrefix    string
}

// BuildManifests returns ready to deploy YAML manifest bundle.
func BuildManifests(r release.ReleaseBundle, opts ...ManifestBuilderOption) (string, error) {
	v := ManifestBuilder{releaseBundle: r}
	for _, opt := range opts {
		opt(&v)
	}
	return v.compile()
}

func WithManifestNamePrefix(prefix string) ManifestBuilderOption {
	return func(b *ManifestBuilder) {
		b.kustomization.namePrefix = prefix + "-"
	}
}

func WithKustomizePatch(patchYAML string) ManifestBuilderOption {
	return func(b *ManifestBuilder) {
		b.kustomization.patches = append(b.kustomization.patches, patchYAML)
	}
}

func WithReplicasPatch(replicas int) ManifestBuilderOption {
	return func(b *ManifestBuilder) {
		WithKustomizePatch(fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
spec:
  replicas: %d`, b.releaseBundle.DeploymentResourceName, replicas))(b)
	}
}

func (b ManifestBuilder) compile() (string, error) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "manifest-staging-")
	if err != nil {
		return "", fmt.Errorf("failed to create staging dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// extract manifests
	if err := b.releaseBundle.WriteTo(tmpDir); err != nil {
		return "", fmt.Errorf("failed to extract manifests: %w", err)
	}

	kf, err := os.OpenFile(filepath.Join(tmpDir, "kustomization.yaml"), os.O_APPEND|os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open kustomization file: %w", err)
	}
	defer kf.Close()

	// write kustomizations
	if b.kustomization.namePrefix != "" {
		if _, err := fmt.Fprintf(kf, "namePrefix: %v\n", b.kustomization.namePrefix); err != nil {
			return "", err
		}
	}

	// write kustomize patches
	patchFileFmt := `patch-%d.yaml`
	if len(b.kustomization.patches) > 0 {
		if _, err := fmt.Fprintln(kf, "patchesStrategicMerge:"); err != nil {
			return "", err
		}
	}
	for patchIndex, patch := range b.kustomization.patches {
		patchFile := filepath.Join(tmpDir, fmt.Sprintf(patchFileFmt, patchIndex))
		if err := os.WriteFile(patchFile, []byte(patch), 0644); err != nil {
			return "", err
		}
		if _, err := fmt.Fprintf(kf, "- %s", patchFile); err != nil {
			return "", err
		}
	}
	if err := kf.Close(); err != nil {
		return "", err
	}

	return kustomizeAt(tmpDir)
}

func kustomizeAt(dir string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("kubectl", "kustomize")
	cmd.Dir = dir
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to kustomize: %w -- stderr=%s", err, stderr.String())
	}
	return stdout.String(), nil
}

// deploy(manifest, inventory)
//	tempdir
//	write manifest
//	write inventory.yaml
//	kpt live apply
//	> output

// destroy(inventory)
//	tempdir
//	write inventory.yaml
//	kpt live destroy
//	> output / error

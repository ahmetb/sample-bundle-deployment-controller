// Copyright 2022 Ahmet Alp Balkan

package deployer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	rgfilev1alpha1 "github.com/GoogleContainerTools/kpt/pkg/api/resourcegroup/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func Deploy(ctx context.Context, inventoryObject client.ObjectKey, inventoryID, manifests string) (string, error) {
	// TODO target different clusters if needed
	// requires significant refactoring in kubernetes-sigs/cli-utils pkg.
	var b bytes.Buffer
	b.WriteString(manifests)
	b.WriteString("\n---\n")
	b.Write(resourceGroupManifest(inventoryObject, inventoryID))
	return kptApply(ctx, b.Bytes())
}

func Destroy(ctx context.Context, inventoryObject client.ObjectKey, inventoryID string) (string, error) {
	return kptDestroy(ctx, resourceGroupManifest(inventoryObject, inventoryID))
}

func kptApply(ctx context.Context, manifests []byte) (string, error) {
	return kpt(ctx, []string{"live", "apply", "-"},
		bytes.NewReader(manifests))
}

func kptDestroy(ctx context.Context, resourceGroupManifest []byte) (string, error) {
	return kpt(ctx, []string{"live", "destroy", "-"},
		bytes.NewReader(resourceGroupManifest))
}

func kpt(ctx context.Context, args []string, in io.Reader) (string, error) {
	cmd := exec.CommandContext(ctx, "kpt", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = in
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("kpt failed: %w -- stderr=%s", err, stderr.String())
	}
	return stdout.String(), nil
}

func resourceGroupManifest(inventoryObject client.ObjectKey, inventoryID string) []byte {
	rg := &rgfilev1alpha1.ResourceGroup{
		ResourceMeta: rgfilev1alpha1.DefaultMeta,
	}
	rg.Name = inventoryObject.Name
	rg.Namespace = inventoryObject.Namespace
	rg.Labels = map[string]string{rgfilev1alpha1.RGInventoryIDLabel: inventoryID}

	rgManifest, _ := yaml.Marshal(rg)
	return rgManifest
}

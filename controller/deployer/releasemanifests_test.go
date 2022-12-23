// Copyright 2022 Ahmet Alp Balkan

package deployer

import (
	"bundle-deployment-controller/controller/release"
	"testing"
)

func TestManifestBuilder(t *testing.T) {
	out, err := BuildManifests(release.ReleaseBundles["rapid"],
		WithManifestNamePrefix("foo-bar-"),
		WithReplicasPatch(999),
	)
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("no out")
	}
	t.Log(out)
}

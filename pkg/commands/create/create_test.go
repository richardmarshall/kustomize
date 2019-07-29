// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/commands/kustfile"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

func readKustomizationFS(t *testing.T, fakeFS fs.FileSystem) *types.Kustomization {
	kf, err := kustfile.NewKustomizationFile(fakeFS)
	if err != nil {
		t.Errorf("unexpected new error %v", err)
	}
	m, err := kf.Read()
	if err != nil {
		t.Errorf("unexpected read error %v", err)
	}
	return m
}
func TestCreateNoArgs(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	cmd := NewCmdCreate(fakeFS)
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	readKustomizationFS(t, fakeFS)
}

func TestCreateWithResources(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	fakeFS.WriteFile("foo.yaml", []byte(""))
	opts := createFlags{resources: []string{"foo.yaml"}}
	err := runCreate(opts, fakeFS)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fakeFS)
	if len(m.Resources) == 0 {
		t.Errorf("resources slice is empty.")
	}
}

func TestCreateWithLabels(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	opts := createFlags{labels: []string{"foo:bar"}}
	err := runCreate(opts, fakeFS)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fakeFS)
	v, found := m.CommonLabels["foo"]
	if !found {
		t.Errorf("expected common label to be set")
	}
	if v != "bar" {
		t.Errorf("want: bar, got: %s", v)
	}
}

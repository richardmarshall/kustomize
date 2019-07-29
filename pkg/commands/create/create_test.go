// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"reflect"
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
	expected := []string{"foo.yaml"}
	if !reflect.DeepEqual(m.Resources, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.Resources)
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
	expected := map[string]string{"foo": "bar"}
	if !reflect.DeepEqual(m.CommonLabels, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.CommonLabels)
	}
}

func TestCreateWithAnnotations(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	opts := createFlags{annotations: []string{"foo:bar"}}
	err := runCreate(opts, fakeFS)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fakeFS)
	expected := map[string]string{"foo": "bar"}
	if !reflect.DeepEqual(m.CommonAnnotations, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.CommonAnnotations)
	}
}

func TestCreateWithNamePrefix(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	want := "foo-"
	opts := createFlags{prefix: want}
	err := runCreate(opts, fakeFS)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fakeFS)
	got := m.NamePrefix
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestCreateWithNameSuffix(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	opts := createFlags{suffix: "-foo"}
	err := runCreate(opts, fakeFS)
	if err != nil {
		t.Errorf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fakeFS)
	if m.NameSuffix != "-foo" {
		t.Errorf("want: -foo, got: %s", m.NameSuffix)
	}
}

func writeDetectContent(fakeFS fs.FileSystem) {
	fakeFS.WriteFile("/test.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test`))
	fakeFS.WriteFile("/README.md", []byte(`
# Not a k8s resource
This file is not a valid kubernetes object.`))
	fakeFS.Mkdir("/sub")
	fakeFS.WriteFile("/sub/test.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test2`))
	fakeFS.WriteFile("/sub/README.md", []byte(`
# Not a k8s resource
This file in a subdirectory is not a valid kubernetes object.`))
	fakeFS.Mkdir("/overlay")
	fakeFS.WriteFile("/overlay/test.yaml", []byte(`
apiVersion: v1
kind: Service
metadata:
  name: test3`))
	fakeFS.WriteFile("/overlay/kustomization.yaml", []byte(`
resources:
- test.yaml`))
}

func TestCreateWithDetect(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	writeDetectContent(fakeFS)
	opts := createFlags{path: "/", detectResources: true}
	err := runCreate(opts, fakeFS)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fakeFS)
	expected := []string{"/test.yaml"}
	if !reflect.DeepEqual(m.Resources, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.Resources)
	}
}

func TestCreateWithDetectRecursive(t *testing.T) {
	fakeFS := fs.MakeFakeFS()
	writeDetectContent(fakeFS)
	opts := createFlags{path: "/", detectResources: true, detectRecursive: true}
	err := runCreate(opts, fakeFS)
	if err != nil {
		t.Fatalf("unexpected cmd error: %v", err)
	}
	m := readKustomizationFS(t, fakeFS)
	expected := []string{"/overlay", "/sub/test.yaml", "/test.yaml"}
	if !reflect.DeepEqual(m.Resources, expected) {
		t.Fatalf("expected %+v but got %+v", expected, m.Resources)
	}
}

// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kustomize/v3/pkg/commands/kustfile"
	"sigs.k8s.io/kustomize/v3/pkg/commands/util"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

type createFlags struct {
	resources       []string
	annotations     []string
	labels          []string
	prefix          string
	suffix          string
	detectResources bool
	detectRecursive bool
}

// NewCmdCreate returns an instance of 'create' subcommand.
func NewCmdCreate(fSys fs.FileSystem) *cobra.Command {
	var opts createFlags
	c := &cobra.Command{
		Use:   "create",
		Short: "Create a new kustomization in the current directory",
		Long:  "",
		Example: `
	# Create a new overlay from the base '../base".
	kustomize create --resource ../base

	# Detect k8s resources in the current directory and generate a kustomization.
	kustomize create --autodetect

	# Set 
	kustomize create --resource depoyment.yaml --resource service.yaml --namespace staging --nameprefix acme-
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts, fSys)
		},
	}
	c.Flags().StringSliceVar(
		&opts.resources,
		"resource",
		[]string{},
		"Name of a file containing a file to add to the kustomization file.")
	c.Flags().StringSliceVar(
		&opts.annotations,
		"annotation",
		[]string{},
		"Add one or more common annotations.")
	c.Flags().StringSliceVar(
		&opts.labels,
		"label",
		[]string{},
		"Add one or more common labels.")
	c.Flags().StringVar(
		&opts.prefix,
		"nameprefix",
		"",
		"Sets the value of the namePrefix field in the kustomization file.")
	c.Flags().StringVar(
		&opts.suffix,
		"namesuffix",
		"",
		"Sets the value of the nameSuffix field in the kustomization file.")
	c.Flags().BoolVar(
		&opts.detectResources,
		"autodetect",
		false,
		"Search for kubernetes resources in the current directory to be added to the kustomization file.")
	c.Flags().BoolVar(
		&opts.detectRecursive,
		"recursive",
		false,
		"Enable recursive directory searching for resource auto-detection.")
	return c
}

func runCreate(opts createFlags, fSys fs.FileSystem) error {
	resources, err := util.GlobPatterns(fSys, opts.resources)
	if err != nil {
		return err
	}
	if _, err := kustfile.NewKustomizationFile(fSys); err == nil {
		return fmt.Errorf("kustomization file already exists")
	}
	if opts.detectResources {
		detected, err := util.DetectResources(fSys, ".", opts.detectRecursive)
		if err != nil {
			return err
		}
		for _, resource := range detected {
			if kustfile.StringInSlice(resource, resources) {
				continue
			}
			resources = append(resources, resource)
		}
	}
	f, err := fSys.Create("kustomization.yaml")
	if err != nil {
		return err
	}
	f.Close()
	mf, err := kustfile.NewKustomizationFile(fSys)
	if err != nil {
		return err
	}
	m, err := mf.Read()
	if err != nil {
		return err
	}
	m.Resources = resources
	m.NamePrefix = opts.prefix
	m.NameSuffix = opts.suffix
	annotations, err := util.ConvertToMap(opts.annotations, "annotation")
	if err != nil {
		return err
	}
	m.CommonAnnotations = annotations
	labels, err := util.ConvertToMap(opts.labels, "label")
	if err != nil {
		return err
	}
	m.CommonLabels = labels
	return mf.Write(m)
}

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
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(opts, fSys)
		},
	}
	c.Flags().StringSliceVar(&opts.resources, "resource", []string{}, "")
	c.Flags().StringSliceVar(&opts.annotations, "annotation", []string{}, "")
	c.Flags().StringSliceVar(&opts.labels, "label", []string{}, "")
	c.Flags().StringVar(&opts.prefix, "nameprefix", "", "")
	c.Flags().StringVar(&opts.suffix, "namesuffix", "", "")
	c.Flags().BoolVar(&opts.detectResources, "autodetect", false, "")
	c.Flags().BoolVar(&opts.detectRecursive, "recursive", false, "")
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
	if opts.detectResources {
		detected, err := util.DetectResources(fSys, ".", opts.detectRecursive)
		if err != nil {
			return err
		}
		for _, resource := range detected {
			if kustfile.StringInSlice(resource, m.Resources) {
				continue
			}
			m.Resources = append(m.Resources, resource)
		}
	}
	m.NamePrefix = opts.prefix
	m.NameSuffix = opts.suffix
	annotations, err := util.ConvertToMap(opts.annotations)
	if err != nil {
		return err
	}
	m.CommonAnnotations = annotations
	labels, err := util.ConvertToMap(opts.labels)
	if err != nil {
		return err
	}
	m.CommonLabels = labels
	return mf.Write(m)
}

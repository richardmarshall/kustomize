// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"log"
	"strings"

	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/pgmconfig"
)

func GlobPatterns(fsys fs.FileSystem, patterns []string) ([]string, error) {
	var result []string
	for _, pattern := range patterns {
		files, err := fsys.Glob(pattern)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			log.Printf("%s has no match", pattern)
			continue
		}
		result = append(result, files...)
	}
	return result, nil
}

func ConvertToMap(inputs []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, input := range inputs {
		c := strings.Index(input, ":")
		if c == 0 {
			// key is not passed
			return nil, fmt.Errorf("invalid %s, %s", input, "need k:v pair where v may be quoted")
		} else if c < 0 {
			// only key passed
			result[input] = ""
		} else {
			// both key and value passed
			key := input[:c]
			value := trimQuotes(input[c+1:])
			result[key] = value
		}
	}
	return result, nil
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func DetectResources(fSys fs.FileSystem, base string, recursive bool) ([]string, error) {
	var paths []string
	factory := kunstruct.NewKunstructuredFactoryImpl()
	err := fSys.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		if info.IsDir() {
			if !recursive {
				return filepath.SkipDir
			}
			for _, kfilename := range pgmconfig.KustomizationFileNames {
				if fSys.Exists(filepath.Join(path, kfilename)) {
					paths = append(paths, path)
					return filepath.SkipDir
				}
			}
			return nil
		}
		fContents, err := fSys.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := factory.SliceFromBytes(fContents); err != nil {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths, err
}

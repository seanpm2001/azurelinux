// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package systemdependency

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/microsoft/azurelinux/toolkit/tools/internal/file"
	"github.com/microsoft/azurelinux/toolkit/tools/internal/shell"
	"github.com/microsoft/azurelinux/toolkit/tools/internal/versioncompare"
)

var (
	// Parses the kernel version from "uname -r" or subdirectories of /lib/modules.
	//
	// Examples:
	//   OS               Version
	//   Fedora 40        6.11.6-200.fc40.x86_64
	//   Ubuntu 22.04     6.8.0-48-generic
	//   Azure Linux 2.0  5.15.153.1-2.cm2
	//   Azure Linux 3.0  6.6.47.1-1.azl3
	kernelVersionRegex = regexp.MustCompile(`^(\d+\.\d+\.\d+)([.\-][a-zA-Z0-9_.\-]*)?$`)
)

func GetBuildHostKernelVersion() (*versioncompare.TolerantVersion, error) {
	stdout, _, err := shell.Execute("uname", "-r")
	if err != nil {
		return nil, fmt.Errorf("failed to get kernel version using uname:\n%w", err)
	}

	stdout = strings.TrimSpace(stdout)

	version, err := parseKernelVersion(stdout)
	if err != nil {
		return nil, err
	}

	return version, nil
}

func GetOldestInstalledKernelVersion(rootfs string) (*versioncompare.TolerantVersion, error) {
	versions, err := GetInstalledKernelVersions(rootfs)
	if err != nil {
		return nil, err
	}

	if len(versions) <= 0 {
		return nil, fmt.Errorf("no installed kernels found in image")
	}

	oldestVersion := versions[0]
	for _, version := range versions {
		if version.Compare(oldestVersion) < 0 {
			oldestVersion = version
		}
	}

	return oldestVersion, nil
}

func GetInstalledKernelVersions(rootfs string) ([]*versioncompare.TolerantVersion, error) {
	versionStrings, err := GetInstalledKernelStringVersions(rootfs)
	if err != nil {
		return nil, fmt.Errorf("failed to get kernel version using uname:\n%w", err)
	}

	versions := []*versioncompare.TolerantVersion(nil)
	for _, versionString := range versionStrings {
		version, err := parseKernelVersion(versionString)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}

func GetInstalledKernelStringVersions(rootfs string) ([]string, error) {
	kernelParentPath := filepath.Join(rootfs, "/lib/modules")
	kernelDirs, err := os.ReadDir(kernelParentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate kernels under (%s):\n%w", kernelParentPath, err)
	}

	// Filter out directories that are empty.
	// Some versions of Azure Linux 2.0 don't cleanup properly when the kernel package is uninstalled.
	filteredKernelDirs := []string(nil)
	for _, kernelDir := range kernelDirs {
		kernelPath := filepath.Join(kernelParentPath, kernelDir.Name())
		empty, err := file.IsDirEmpty(kernelPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check if directory (%s) is empty:\n%w", kernelPath, err)
		}

		if !empty {
			filteredKernelDirs = append(filteredKernelDirs, kernelDir.Name())
		}
	}

	return filteredKernelDirs, nil
}

func parseKernelVersion(versionString string) (*versioncompare.TolerantVersion, error) {
	match := kernelVersionRegex.FindStringSubmatch(versionString)
	if match == nil {
		return nil, fmt.Errorf("failed to parse kernel version (%s)", versionString)
	}

	majorMinorPatchString := match[1]
	majorMinorPatch := versioncompare.New(majorMinorPatchString)
	return majorMinorPatch, nil
}

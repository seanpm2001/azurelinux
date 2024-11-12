// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package imagecustomizerlib

import (
	"fmt"

	"github.com/microsoft/azurelinux/toolkit/tools/internal/safechroot"
	"github.com/microsoft/azurelinux/toolkit/tools/internal/systemdependency"
)

// Check if the user accidentally uninstalled the kernel package without installing a substitute package.
func checkForInstalledKernel(imageChroot *safechroot.Chroot) error {
	kernels, err := systemdependency.GetInstalledKernelStringVersions(imageChroot.RootDir())
	if err != nil {
		return err
	}

	if len(kernels) <= 0 {
		return fmt.Errorf("no installed kernel found")
	}

	return nil
}

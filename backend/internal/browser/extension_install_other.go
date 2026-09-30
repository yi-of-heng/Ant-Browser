//go:build !windows
// +build !windows

package browser

// Stage a per-profile copy for Chromium's unpacked extension loader. There is
// no helper browser process and no attempt to forge protected preferences.
func installCRXIntoProfile(userDataDir string, _ string, packagePath string, extension Extension, _ []string) (string, error) {
	return installExtensionPackageIntoProfileWithRegistration(userDataDir, packagePath, extension, false)
}

func browserExtensionInstallerExitHint(_ string) string {
	return ""
}

func (m *Manager) cleanupManagedExternalExtensionRegistry() error {
	return nil
}

func ensurePersistentExternalExtensionRegistry(runtimeID string, packagePath string, version string) error {
	return nil
}

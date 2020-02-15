package getproviders

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform/addrs"
)

// FilesystemMirrorSource is a source that reads providers and their metadata
// from a directory prefix in the local filesystem.
type FilesystemMirrorSource struct {
	baseDir string
}

var _ Source = (*FilesystemMirrorSource)(nil)

// NewFilesystemMirrorSource constructs and returns a new filesystem-based
// mirror source with the given base directory.
func NewFilesystemMirrorSource(baseDir string) *FilesystemMirrorSource {
	return &FilesystemMirrorSource{
		baseDir: baseDir,
	}
}

// AvailableVersions scans the directory structure under the source's base
// directory for locally-mirrored packages for the given provider, returning
// a list of version numbers for the providers it found.
func (s *FilesystemMirrorSource) AvailableVersions(provider addrs.Provider) (VersionList, error) {
	// TODO: Implement
	panic("FilesystemMirrorSource.AvailableVersions not yet implemented")
}

// PackageMeta checks to see if the source's base directory contains a
// local copy of the distribution package for the given provider version on
// the given target, and returns the metadata about it if so.
func (s *FilesystemMirrorSource) PackageMeta(provider addrs.Provider, version Version, target Platform) (PackageMeta, error) {
	// TODO: Implement
	panic("FilesystemMirrorSource.PackageMeta not yet implemented")
}

// AllAvailablePackages scans the directory structure under the source's base
// directory for locally-mirrored packages for all providers, returning a map
// of the discovered packages with the fully-qualified provider names as
// keys.
//
// This is not an operation generally supported by all Source implementations,
// but the filesystem implementation offers it because we also use the
// filesystem mirror source directly to scan our auto-install plugin directory
// and in other automatic discovery situations.
func (s *FilesystemMirrorSource) AllAvailablePackages() (map[addrs.Provider][]PackageMeta, error) {
	ret := make(map[addrs.Provider][]PackageMeta)
	err := filepath.Walk(s.baseDir, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("cannot search %s: %s", fullPath, err)
		}

		// There are two valid directory structures that we support here...
		// Unpacked: registry.terraform.io/hashicorp/aws/2.0.0/linux_amd64 (a directory)
		// Packed:   registry.terraform.io/hashicorp/aws/terraform-provider-aws_2.0.0_linux_amd64.zip (a file)
		//
		// Both of these give us enough information to identify the package
		// metadata.
		fsPath, err := filepath.Rel(s.baseDir, fullPath)
		if err != nil {
			// This should never happen because the filepath.Walk contract is
			// for the paths to include the base path.
			log.Printf("[TRACE] FilesystemMirrorSource: ignoring malformed path %q during walk: %s", fullPath, err)
			return nil
		}
		relPath := filepath.ToSlash(fsPath)
		parts := strings.Split(relPath, "/")

		if len(parts) < 3 {
			// Likely a prefix of a valid path, so we'll ignore it and visit
			// the full valid path on a later call.
			return nil
		}

		hostnameGiven := parts[0]
		namespace := parts[1]
		typeName := parts[2]

		hostname, err := svchost.ForComparison(hostnameGiven)
		if err != nil {
			log.Printf("[WARN] local provider path %q contains invalid hostname %q; ignoring", fullPath, hostnameGiven)
			return nil
		}
		var providerAddr addrs.Provider
		if namespace == addrs.LegacyProviderNamespace {
			if hostname != addrs.DefaultRegistryHost {
				log.Printf("[WARN] local provider path %q indicates a legacy provider not on the default registry host; ignoring", fullPath)
				return nil
			}
			providerAddr = addrs.NewLegacyProvider(typeName)
		} else {
			providerAddr = addrs.NewProvider(hostname, namespace, typeName)
		}

		switch len(parts) {
		case 5: // Might be unpacked layout
			if !info.IsDir() {
				return nil // packed layout requires a directory
			}

			versionStr := parts[3]
			version, err := ParseVersion(versionStr)
			if err != nil {
				log.Printf("[WARN] ignoring local provider path %q with invalid version %q: %s", fullPath, versionStr, err)
				return nil
			}

			platformStr := parts[4]
			platform, err := ParsePlatform(platformStr)
			if err != nil {
				log.Printf("[WARN] ignoring local provider path %q with invalid platform %q: %s", fullPath, platformStr, err)
				return nil
			}

			log.Printf("[TRACE] FilesystemMirrorSource: found %s v%s for %s at %s", providerAddr, version, platform, fullPath)

			meta := PackageMeta{
				// FIXME: Should we include the provider address and version
				// number in here too? At least one of them is often redundant
				// due to being embedded in another data structure, but
				// often not both of them.

				// FIXME: How do we populate this?
				ProtocolVersions: nil,
				TargetPlatform:   platform,

				// Because this is already unpacked, the filename is synthetic
				// based on the standard naming scheme.
				Filename: fmt.Sprintf("terraform-provider-%s_%s_%s.zip", providerAddr.Type, version, platform),
				Location: PackageLocalDir(fullPath),

				// FIXME: What about the SHA256Sum field? As currently specified
				// it's a hash of the zip file, but this thing is already
				// unpacked and so we don't have the zip file to hash.
			}
			ret[providerAddr] = append(ret[providerAddr], meta)

		case 4: // Might be packed layout
			if info.IsDir() {
				return nil // packed layout requires a file
			}

			log.Printf("[TRACE] FilesystemMirrorSource: found %s (packed) at %s", providerAddr, fullPath)

		}

		return nil
	})
	return ret, err
}

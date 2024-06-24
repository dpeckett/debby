package packages

import (
	"sync"

	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/types/version"

	"github.com/google/btree"
)

// PackageCollection is a collection of packages.
type PackageCollection struct {
	mu   sync.RWMutex
	tree *btree.BTreeG[types.Package]
}

// NewPackageCollection creates a new collection of packages.
func NewPackageCollection() *PackageCollection {
	return &PackageCollection{
		tree: btree.NewG(2, func(a, b types.Package) bool { // TODO: tune the degree.
			return a.Compare(b) < 0
		}),
	}
}

// Len returns the number of packages in the collection.
func (pc *PackageCollection) Len() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	return pc.tree.Len()
}

// Add adds a package to the collection.
func (pc *PackageCollection) Add(pkg types.Package) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.tree.ReplaceOrInsert(pkg)
}

// AddAll adds multiple packages to the collection.
func (pc *PackageCollection) AddAll(packages []types.Package) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for _, pkg := range packages {
		pc.tree.ReplaceOrInsert(pkg)
	}
}

// Get returns all packages that match the provided name.
func (pc *PackageCollection) Get(name string) (packages []types.Package) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	pc.tree.AscendGreaterOrEqual(types.Package{
		Package: name,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		packages = append(packages, pkg)

		return true
	})
	return
}

// StrictlyEarlier returns all packages that match the provided name and are
// strictly earlier than the provided version.
func (pc *PackageCollection) StrictlyEarlier(name string, version version.Version) (packages []types.Package) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	pc.tree.DescendLessOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		// Skip the package if it is the same version (since we want strictly earlier)
		if pkg.Version.Compare(version) == 0 {
			return true
		}

		packages = append(packages, pkg)

		return true
	})
	return
}

// EarlierOrEqual returns all packages that match the provided name and are
// earlier or equal to the provided version.
func (pc *PackageCollection) EarlierOrEqual(name string, version version.Version) (packages []types.Package) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	pc.tree.DescendLessOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		packages = append(packages, pkg)

		return true
	})
	return
}

// ExactlyEqual returns the package that matches the provided name and version.
func (pc *PackageCollection) ExactlyEqual(name string, version version.Version) (*types.Package, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	var foundPackage *types.Package
	pc.tree.AscendGreaterOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		if pkg.Version.Compare(version) == 0 {
			foundPackage = &pkg
		}

		return false
	})
	return foundPackage, foundPackage != nil
}

// LaterOrEqual returns all packages that match the provided name and are
// later or equal to the provided version.
func (pc *PackageCollection) LaterOrEqual(name string, version version.Version) (packages []types.Package) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	pc.tree.AscendGreaterOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		packages = append(packages, pkg)

		return true
	})
	return
}

// StrictlyLater returns all packages that match the provided name and are
// strictly later than the provided version.
func (pc *PackageCollection) StrictlyLater(name string, version version.Version) (packages []types.Package) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	pc.tree.AscendGreaterOrEqual(types.Package{
		Package: name,
		Version: version,
	}, func(pkg types.Package) bool {
		if pkg.Package != name {
			return false
		}

		// Skip the package if it is the same version (since we want strictly later)
		if pkg.Version.Compare(version) == 0 {
			return true
		}

		packages = append(packages, pkg)

		return true
	})
	return
}

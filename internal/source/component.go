package source

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/util"
	"github.com/ulikunitz/xz"
)

// Component represents a component of a Debian repository.
type Component struct {
	// Name is the name of the component.
	Name string
	// PackagesURL is the URL to the Packages file for the component.
	PackagesURL *url.URL
	// PackagesSHA256Sum is the SHA256 sum of the Packages file.
	PackagesSHA256Sum []byte
	// Internal fields.
	httpClient *http.Client
	keyring    openpgp.EntityList
	sourceURL  *url.URL
}

func (c *Component) Packages(ctx context.Context) ([]types.Package, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.PackagesURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download Packages file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download Packages file: %s", resp.Status)
	}

	hr := util.NewHashReader(resp.Body)

	xzReader, err := xz.NewReader(hr)
	if err != nil {
		return nil, fmt.Errorf("failed to create xz reader: %w", err)
	}

	decoder, err := deb822.NewDecoder(xzReader, c.keyring)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	var packages []types.Package
	if err := decoder.Decode(&packages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Packages file: %w", err)
	}

	if err := hr.Verify(c.PackagesSHA256Sum); err != nil {
		return nil, fmt.Errorf("failed to verify Packages file: %w", err)
	}

	packageURL, err := url.Parse(c.sourceURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URL: %w", err)
	}
	basePath := packageURL.Path

	for i := range packages {
		packageURL.Path = path.Join(basePath, packages[i].Filename)
		packages[i].URLs = append(packages[i].URLs, packageURL.String())
	}

	return packages, nil
}

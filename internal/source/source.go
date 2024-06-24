package source

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"

	"github.com/ProtonMail/go-crypto/openpgp"
	latestconfig "github.com/dpeckett/debby/internal/config/v1alpha1"
	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/types/arch"
)

const (
	defaultDistribution = "stable"
)

var defaultComponents = []string{"main"}

// Source represents a Debian repository source.
type Source struct {
	logger       *slog.Logger
	httpClient   *http.Client
	keyring      openpgp.EntityList
	sourceURL    *url.URL
	distribution string
	components   []string
}

// NewSource creates a new Debian repository source.
func NewSource(logger *slog.Logger, httpClient *http.Client, keyring openpgp.EntityList, conf latestconfig.SourceConfig) (*Source, error) {
	distribution := defaultDistribution
	if conf.Distribution != "" {
		distribution = conf.Distribution
	}

	components := defaultComponents
	if len(conf.Components) > 0 {
		components = conf.Components
	}

	sourceURL, err := url.Parse(conf.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URL: %w", err)
	}

	return &Source{
		logger: logger.With(
			slog.String("sourceURL", conf.URL),
			slog.String("distribution", distribution)),
		httpClient:   httpClient,
		keyring:      keyring,
		sourceURL:    sourceURL,
		distribution: distribution,
		components:   components,
	}, nil
}

// Components returns the components available in the source for the target architecture.
func (s *Source) Components(ctx context.Context, targetArch arch.Arch) ([]Component, error) {
	inReleaseURL, err := url.Parse(s.sourceURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URL: %w", err)
	}

	inReleaseURL.Path = path.Join(inReleaseURL.Path, "dists", s.distribution, "InRelease")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, inReleaseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download InRelease file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download InRelease file: %s", resp.Status)
	}

	decoder, err := deb822.NewDecoder(resp.Body, s.keyring)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	if decoder.Signer() == nil {
		return nil, errors.New("InRelease file is not signed")
	}

	var release types.Release
	if err := decoder.Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to unmarshal InRelease file: %w", err)
	}

	allArch := arch.MustParse("all")
	var availableArchitectures []arch.Arch
	for _, releaseArch := range release.Architectures {
		if releaseArch.Is(&allArch) || releaseArch.Is(&targetArch) {
			availableArchitectures = append(availableArchitectures, releaseArch)
		}
	}

	if len(availableArchitectures) == 0 {
		s.logger.Warn("No architectures available")
		return nil, nil
	}

	desiredComponents := map[string]bool{}
	for _, component := range defaultComponents {
		desiredComponents[component] = true
	}
	for _, component := range s.components {
		desiredComponents[component] = true
	}

	var availableComponents []string
	for _, component := range release.Components {
		if desiredComponents[component] {
			availableComponents = append(availableComponents, component)
		}
	}

	if len(availableComponents) == 0 {
		s.logger.Warn("No components available")
		return nil, nil
	}

	// Get the SHA256 sums for files in the release.
	releaseSHA256Sums, err := release.SHA256Sums()
	if err != nil {
		return nil, fmt.Errorf("failed to get SHA256 sums: %w", err)
	}

	var components []Component
	for _, component := range availableComponents {
		for _, arch := range availableArchitectures {
			packagesURL, err := url.Parse(s.sourceURL.String())
			if err != nil {
				return nil, fmt.Errorf("failed to parse source URL: %w", err)
			}

			packagesURL.Path = path.Join(packagesURL.Path, "dists", s.distribution, component, "binary-"+arch.String(), "Packages.xz")

			relativePackagesPath := path.Join(path.Base(component), "binary-"+arch.String(), "Packages.xz")
			expectedSHA256Sum, ok := releaseSHA256Sums[relativePackagesPath]
			if !ok {
				return nil, fmt.Errorf("no SHA256 sum for Packages file: %s", relativePackagesPath)
			}

			components = append(components, Component{
				Name:              component,
				PackagesURL:       packagesURL,
				PackagesSHA256Sum: expectedSHA256Sum,
				httpClient:        s.httpClient,
				keyring:           s.keyring,
				sourceURL:         s.sourceURL,
			})
		}
	}

	return components, nil
}

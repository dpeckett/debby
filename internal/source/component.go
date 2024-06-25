// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

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
	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/types/arch"
	"github.com/dpeckett/debby/internal/util"
)

// Component represents a component of a Debian repository.
type Component struct {
	// Name is the name of the component.
	Name string
	// Arch is the architecture of the component.
	Arch arch.Arch
	// URL is the base URL of the component.
	URL *url.URL
	// SHA256Sums are the SHA256 sums of files in the component.
	SHA256Sums map[string][]byte
	// Internal fields.
	httpClient *http.Client
	keyring    openpgp.EntityList
	sourceURL  *url.URL
}

func (c *Component) Packages(ctx context.Context) ([]types.Package, error) {
	var errs error

	for _, name := range []string{"Packages.xz", "Packages.gz", "Packages"} {
		packagesURL, err := url.Parse(c.URL.String())
		if err != nil {
			return nil, fmt.Errorf("failed to parse component URL: %w", err)
		}

		packagesURL.Path = path.Join(packagesURL.Path, name)

		slog.Debug("Attempting to download Packages file", slog.String("url", packagesURL.String()))

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, packagesURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to download %s file: %w", name, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errs = errors.Join(errs, fmt.Errorf("failed to download %s file: %s", name, resp.Status))
			continue
		}

		hr := util.NewHashReader(resp.Body)

		dr, err := util.Decompress(hr)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to decompress %s file: %w", name, err))
			continue
		}

		decoder, err := deb822.NewDecoder(dr, c.keyring)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to create decoder: %w", err))
			continue
		}

		var packageList []types.Package
		if err := decoder.Decode(&packageList); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to unmarshal %s file: %w", name, err))
			continue
		}

		if err := hr.Verify(c.SHA256Sums[name]); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to verify %s file: %w", name, err))
			continue
		}

		packageURL, err := url.Parse(c.sourceURL.String())
		if err != nil {
			return nil, fmt.Errorf("failed to parse source URL: %w", err)
		}
		basePath := packageURL.Path

		for i := range packageList {
			packageURL.Path = path.Join(basePath, packageList[i].Filename)
			packageList[i].URLs = append(packageList[i].URLs, packageURL.String())
		}

		return packageList, nil
	}

	return nil, fmt.Errorf("failed to download Packages file: %w", errs)
}

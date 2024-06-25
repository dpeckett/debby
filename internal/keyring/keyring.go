// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package keyring

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// Load reads an OpenPGP keyring from a file or URL.
func Load(ctx context.Context, logger *slog.Logger, httpClient *http.Client, key string) (openpgp.EntityList, error) {
	if len(key) == 0 {
		return openpgp.EntityList{}, nil
	}

	// If the key is a URL, download it.
	if strings.Contains(key, "://") {
		logger.Debug("Downloading key", slog.String("url", key))

		keyURL, err := url.Parse(key)
		if err != nil {
			return nil, err
		}

		if keyURL.Scheme != "https" {
			return nil, errors.New("key URL must be HTTPS")
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, keyURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to download key: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download key: %s", resp.Status)
		}

		// ReadArmoredKeyRing() doesn't read the entire response body, so we need
		// to do it ourselves (so that response caching will work as expected).
		keyringData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return openpgp.ReadArmoredKeyRing(bytes.NewReader(keyringData))
	} else { // If the key is a file, open it.
		logger.Debug("Reading key file", slog.String("path", key))

		f, err := os.Open(key)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return openpgp.ReadArmoredKeyRing(f)
	}
}

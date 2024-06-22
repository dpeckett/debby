// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package util

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/rogpeppe/go-internal/cache"
)

// DiskCache is a cache that stores http responses on disk.
type DiskCache struct {
	*cache.Cache
	logger    *slog.Logger
	namespace string
}

// NewDiskCache creates a new cache that stores responses in the given directory.
// The namespace is used to separate different caches in the same directory.
func NewDiskCache(logger *slog.Logger, dir, namespace string) (*DiskCache, error) {
	c, err := cache.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("error opening cache: %w", err)
	}

	if err := c.Trim(); err != nil {
		return nil, fmt.Errorf("error trimming cache: %w", err)
	}

	return &DiskCache{
		Cache:     c,
		logger:    logger,
		namespace: namespace,
	}, nil
}

func (c *DiskCache) Get(key string) ([]byte, bool) {
	responseBytes, _, err := c.Cache.GetBytes(c.getActionID(key))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			c.logger.Warn("Error getting cached response",
				slog.String("key", key), slog.Any("error", err))
		} else {
			c.logger.Debug("Cache miss", slog.String("key", key))
		}

		return nil, false
	}

	c.logger.Debug("Cache hit", slog.String("key", key))

	return responseBytes, true
}

func (c *DiskCache) Set(key string, responseBytes []byte) {
	c.logger.Debug("Storing cached response", slog.String("key", key))

	if err := c.Cache.PutBytes(c.getActionID(key), responseBytes); err != nil {
		c.logger.Warn("Error setting cached response", slog.Any("error", err))
	}
}

func (c *DiskCache) Delete(key string) {}

func (c *DiskCache) getActionID(key string) cache.ActionID {
	h := cache.NewHash(c.namespace)
	_, _ = h.Write([]byte(key))
	return h.Sum()
}

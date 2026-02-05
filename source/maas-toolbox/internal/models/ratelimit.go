/*
 * This source file includes portions generated or suggested by
 * artificial intelligence tools and subsequently reviewed,
 * modified, and validated by human contributors.
 *
 * Human authorship, design decisions, and final responsibility
 * for this code remain with the project contributors.
 */

// Copyright 2025 Bryon Baker
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

// RateLimit represents a single rate limit definition within a RateLimitPolicy
// @Description Rate limit configuration that defines request count limits per tier
type RateLimit struct {
	Name   string `json:"name" yaml:"name" example:"free"`       // Limit name (unique within policy)
	Limit  int    `json:"limit" yaml:"limit" example:"5"`        // Request count limit
	Window string `json:"window" yaml:"window" example:"2m"`     // Time window (Kubernetes duration format)
	Tier   string `json:"tier" yaml:"tier" example:"free"`       // Tier name for predicate
}

// RateLimitUpdate represents fields that can be updated for an existing rate limit
// @Description Update request for rate limit (only limit and window can be changed)
type RateLimitUpdate struct {
	Limit  int    `json:"limit" yaml:"limit" example:"10"`   // Request count limit
	Window string `json:"window" yaml:"window" example:"3m"` // Time window (Kubernetes duration format)
}

// Validate validates a RateLimit struct
func (rl *RateLimit) Validate() error {
	if rl.Name == "" {
		return ErrRateLimitNameRequired
	}
	if rl.Limit <= 0 {
		return ErrRateLimitInvalid
	}
	if rl.Window == "" {
		return ErrRateLimitWindowRequired
	}
	// Validate window format (Kubernetes duration: 1s, 1m, 1h, etc.)
	if !isValidDuration(rl.Window) {
		return ErrRateLimitWindowInvalid
	}
	if rl.Tier == "" {
		return ErrRateLimitTierRequired
	}
	return nil
}

// IsValid returns true if the rate limit is valid
func (rl *RateLimit) IsValid() bool {
	return rl.Validate() == nil
}

// Validate validates a RateLimitUpdate struct
func (rlu *RateLimitUpdate) Validate() error {
	if rlu.Limit <= 0 {
		return ErrRateLimitInvalid
	}
	if rlu.Window == "" {
		return ErrRateLimitWindowRequired
	}
	// Validate window format (Kubernetes duration: 1s, 1m, 1h, etc.)
	if !isValidDuration(rlu.Window) {
		return ErrRateLimitWindowInvalid
	}
	return nil
}

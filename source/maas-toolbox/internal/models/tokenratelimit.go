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

import "regexp"

// TokenRateLimit represents a single token rate limit definition within a TokenRateLimitPolicy
// @Description Token rate limit configuration that defines token consumption limits per tier
type TokenRateLimit struct {
	Name   string `json:"name" yaml:"name" example:"free-user-tokens"`       // Limit name (unique within policy)
	Limit  int    `json:"limit" yaml:"limit" example:"100"`                  // Token count limit
	Window string `json:"window" yaml:"window" example:"1m"`                 // Time window (Kubernetes duration format)
	Tier   string `json:"tier" yaml:"tier" example:"free"`                   // Tier name for predicate
}

// Validate validates a TokenRateLimit struct
func (trl *TokenRateLimit) Validate() error {
	if trl.Name == "" {
		return ErrTokenRateLimitNameRequired
	}
	if trl.Limit <= 0 {
		return ErrTokenRateLimitInvalid
	}
	if trl.Window == "" {
		return ErrTokenRateLimitWindowRequired
	}
	// Validate window format (Kubernetes duration: 1s, 1m, 1h, etc.)
	if !isValidDuration(trl.Window) {
		return ErrTokenRateLimitWindowInvalid
	}
	if trl.Tier == "" {
		return ErrTokenRateLimitTierRequired
	}
	return nil
}

// IsValid returns true if the token rate limit is valid
func (trl *TokenRateLimit) IsValid() bool {
	return trl.Validate() == nil
}

// TokenRateLimitUpdate represents fields that can be updated for an existing token rate limit
// @Description Update request for token rate limit (only limit and window can be changed)
type TokenRateLimitUpdate struct {
	Limit  int    `json:"limit" yaml:"limit" example:"200"`   // Token count limit
	Window string `json:"window" yaml:"window" example:"2m"`  // Time window (Kubernetes duration format)
}

// Validate validates a TokenRateLimitUpdate struct
func (trlu *TokenRateLimitUpdate) Validate() error {
	if trlu.Limit <= 0 {
		return ErrTokenRateLimitInvalid
	}
	if trlu.Window == "" {
		return ErrTokenRateLimitWindowRequired
	}
	// Validate window format (Kubernetes duration: 1s, 1m, 1h, etc.)
	if !isValidDuration(trlu.Window) {
		return ErrTokenRateLimitWindowInvalid
	}
	return nil
}

// isValidDuration checks if a string is a valid Kubernetes duration format
// Valid formats: 1s, 30s, 1m, 5m, 1h, 24h, etc.
func isValidDuration(duration string) bool {
	// Match Kubernetes duration format: number followed by s, m, or h
	match, _ := regexp.MatchString(`^[0-9]+[smh]$`, duration)
	return match
}

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

package service

import (
	"maas-toolbox/internal/models"
)

// RateLimitStorage defines the interface for rate limit storage operations
type RateLimitStorage interface {
	GetAllRateLimits() ([]models.RateLimit, error)
	GetRateLimit(name string) (*models.RateLimit, error)
	CreateRateLimit(rl *models.RateLimit) error
	UpdateRateLimit(name string, rl *models.RateLimit) error
	DeleteRateLimit(name string) error
}

// RateLimitService provides business logic for rate limit operations
type RateLimitService struct {
	storage     RateLimitStorage
	tierService *TierService
}

// NewRateLimitService creates a new RateLimitService instance
func NewRateLimitService(storage RateLimitStorage, tierService *TierService) *RateLimitService {
	return &RateLimitService{
		storage:     storage,
		tierService: tierService,
	}
}

// GetAllRateLimits retrieves all rate limits
func (s *RateLimitService) GetAllRateLimits() ([]models.RateLimit, error) {
	return s.storage.GetAllRateLimits()
}

// GetRateLimit retrieves a specific rate limit by name
func (s *RateLimitService) GetRateLimit(name string) (*models.RateLimit, error) {
	return s.storage.GetRateLimit(name)
}

// CreateRateLimit creates a new rate limit with validation
func (s *RateLimitService) CreateRateLimit(rl *models.RateLimit) error {
	// Validate the rate limit
	if err := rl.Validate(); err != nil {
		return err
	}
	
	// Validate tier exists
	if _, err := s.tierService.GetTier(rl.Tier); err != nil {
		return models.ErrTierNotFound
	}
	
	// Create the rate limit
	return s.storage.CreateRateLimit(rl)
}

// UpdateRateLimit updates an existing rate limit with validation
// Only limit and window can be updated; name and tier are immutable
func (s *RateLimitService) UpdateRateLimit(name string, update *models.RateLimitUpdate) error {
	// Validate the update
	if err := update.Validate(); err != nil {
		return err
	}
	
	// Fetch the existing rate limit to preserve immutable fields
	existing, err := s.storage.GetRateLimit(name)
	if err != nil {
		return err
	}
	
	// Create updated rate limit with immutable fields from existing
	updated := &models.RateLimit{
		Name:   existing.Name,  // Immutable
		Tier:   existing.Tier,  // Immutable
		Limit:  update.Limit,   // Mutable
		Window: update.Window,  // Mutable
	}
	
	// Validate the complete updated rate limit
	if err := updated.Validate(); err != nil {
		return err
	}
	
	// Update the rate limit
	return s.storage.UpdateRateLimit(name, updated)
}

// DeleteRateLimit deletes a rate limit
func (s *RateLimitService) DeleteRateLimit(name string) error {
	return s.storage.DeleteRateLimit(name)
}

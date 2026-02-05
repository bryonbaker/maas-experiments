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
	"fmt"
	"maas-toolbox/internal/models"
	"maas-toolbox/internal/storage"
)

// TokenRateLimitService provides business logic for token rate limit management
type TokenRateLimitService struct {
	storage     *storage.K8sTokenRateLimitStorage
	tierService *TierService
}

// NewTokenRateLimitService creates a new TokenRateLimitService instance
func NewTokenRateLimitService(storage *storage.K8sTokenRateLimitStorage, tierService *TierService) *TokenRateLimitService {
	return &TokenRateLimitService{
		storage:     storage,
		tierService: tierService,
	}
}

// CreateTokenRateLimit creates a new token rate limit
func (s *TokenRateLimitService) CreateTokenRateLimit(trl *models.TokenRateLimit) error {
	// Validate token rate limit
	if err := trl.Validate(); err != nil {
		return err
	}
	
	// Validate that the tier exists
	if _, err := s.tierService.GetTier(trl.Tier); err != nil {
		return fmt.Errorf("tier %s not found: %w", trl.Tier, err)
	}
	
	// Create the token rate limit
	if err := s.storage.CreateTokenRateLimit(trl); err != nil {
		return fmt.Errorf("failed to create token rate limit: %w", err)
	}
	
	return nil
}

// GetAllTokenRateLimits returns all token rate limits
func (s *TokenRateLimitService) GetAllTokenRateLimits() ([]models.TokenRateLimit, error) {
	tokenRateLimits, err := s.storage.GetAllTokenRateLimits()
	if err != nil {
		return nil, fmt.Errorf("failed to get token rate limits: %w", err)
	}
	return tokenRateLimits, nil
}

// GetTokenRateLimit returns a specific token rate limit by name
func (s *TokenRateLimitService) GetTokenRateLimit(name string) (*models.TokenRateLimit, error) {
	if name == "" {
		return nil, models.ErrTokenRateLimitNameRequired
	}
	
	tokenRateLimit, err := s.storage.GetTokenRateLimit(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get token rate limit: %w", err)
	}
	
	return tokenRateLimit, nil
}

// UpdateTokenRateLimit updates an existing token rate limit (only limit and window can be changed)
func (s *TokenRateLimitService) UpdateTokenRateLimit(name string, update *models.TokenRateLimitUpdate) error {
	if name == "" {
		return models.ErrTokenRateLimitNameRequired
	}
	
	// Validate update fields
	if err := update.Validate(); err != nil {
		return err
	}
	
	// Get existing token rate limit to preserve name and tier
	existing, err := s.storage.GetTokenRateLimit(name)
	if err != nil {
		return fmt.Errorf("failed to get existing token rate limit: %w", err)
	}
	
	// Create updated token rate limit with only limit and window changed
	updated := &models.TokenRateLimit{
		Name:   existing.Name,   // Preserve existing name
		Tier:   existing.Tier,   // Preserve existing tier
		Limit:  update.Limit,    // Update limit
		Window: update.Window,   // Update window
	}
	
	// Update the token rate limit
	if err := s.storage.UpdateTokenRateLimit(name, updated); err != nil {
		return fmt.Errorf("failed to update token rate limit: %w", err)
	}
	
	return nil
}

// DeleteTokenRateLimit deletes a token rate limit by name
func (s *TokenRateLimitService) DeleteTokenRateLimit(name string) error {
	if name == "" {
		return models.ErrTokenRateLimitNameRequired
	}
	
	// Delete the token rate limit
	if err := s.storage.DeleteTokenRateLimit(name); err != nil {
		return fmt.Errorf("failed to delete token rate limit: %w", err)
	}
	
	return nil
}

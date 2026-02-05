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

import (
	"testing"
)

func TestTokenRateLimitValidate(t *testing.T) {
	tests := []struct {
		name    string
		trl     TokenRateLimit
		wantErr bool
		errType error
	}{
		{
			name: "valid token rate limit",
			trl: TokenRateLimit{
				Name:   "free-user-tokens",
				Limit:  100,
				Window: "1m",
				Tier:   "free",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			trl: TokenRateLimit{
				Limit:  100,
				Window: "1m",
				Tier:   "free",
			},
			wantErr: true,
			errType: ErrTokenRateLimitNameRequired,
		},
		{
			name: "zero limit",
			trl: TokenRateLimit{
				Name:   "free-user-tokens",
				Limit:  0,
				Window: "1m",
				Tier:   "free",
			},
			wantErr: true,
			errType: ErrTokenRateLimitInvalid,
		},
		{
			name: "negative limit",
			trl: TokenRateLimit{
				Name:   "free-user-tokens",
				Limit:  -10,
				Window: "1m",
				Tier:   "free",
			},
			wantErr: true,
			errType: ErrTokenRateLimitInvalid,
		},
		{
			name: "missing window",
			trl: TokenRateLimit{
				Name:  "free-user-tokens",
				Limit: 100,
				Tier:  "free",
			},
			wantErr: true,
			errType: ErrTokenRateLimitWindowRequired,
		},
		{
			name: "invalid window format",
			trl: TokenRateLimit{
				Name:   "free-user-tokens",
				Limit:  100,
				Window: "invalid",
				Tier:   "free",
			},
			wantErr: true,
			errType: ErrTokenRateLimitWindowInvalid,
		},
		{
			name: "valid window formats",
			trl: TokenRateLimit{
				Name:   "premium-user-tokens",
				Limit:  1000,
				Window: "30s",
				Tier:   "premium",
			},
			wantErr: false,
		},
		{
			name: "missing tier",
			trl: TokenRateLimit{
				Name:   "free-user-tokens",
				Limit:  100,
				Window: "1m",
			},
			wantErr: true,
			errType: ErrTokenRateLimitTierRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trl.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenRateLimit.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != tt.errType {
				t.Errorf("TokenRateLimit.Validate() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestTokenRateLimitIsValid(t *testing.T) {
	validTRL := TokenRateLimit{
		Name:   "free-user-tokens",
		Limit:  100,
		Window: "1m",
		Tier:   "free",
	}

	if !validTRL.IsValid() {
		t.Error("Expected valid TokenRateLimit to return true from IsValid()")
	}

	invalidTRL := TokenRateLimit{
		Name:  "invalid",
		Limit: -1,
		Tier:  "free",
	}

	if invalidTRL.IsValid() {
		t.Error("Expected invalid TokenRateLimit to return false from IsValid()")
	}
}

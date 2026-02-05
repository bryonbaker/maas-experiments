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

func TestRateLimitValidate(t *testing.T) {
	tests := []struct {
		name    string
		rl      RateLimit
		wantErr bool
		errType error
	}{
		{
			name: "valid rate limit",
			rl: RateLimit{
				Name:   "free",
				Limit:  5,
				Window: "2m",
				Tier:   "free",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			rl: RateLimit{
				Limit:  5,
				Window: "2m",
				Tier:   "free",
			},
			wantErr: true,
			errType: ErrRateLimitNameRequired,
		},
		{
			name: "zero limit",
			rl: RateLimit{
				Name:   "free",
				Limit:  0,
				Window: "2m",
				Tier:   "free",
			},
			wantErr: true,
			errType: ErrRateLimitInvalid,
		},
		{
			name: "negative limit",
			rl: RateLimit{
				Name:   "free",
				Limit:  -10,
				Window: "2m",
				Tier:   "free",
			},
			wantErr: true,
			errType: ErrRateLimitInvalid,
		},
		{
			name: "missing window",
			rl: RateLimit{
				Name:  "free",
				Limit: 5,
				Tier:  "free",
			},
			wantErr: true,
			errType: ErrRateLimitWindowRequired,
		},
		{
			name: "invalid window format",
			rl: RateLimit{
				Name:   "free",
				Limit:  5,
				Window: "invalid",
				Tier:   "free",
			},
			wantErr: true,
			errType: ErrRateLimitWindowInvalid,
		},
		{
			name: "valid window formats",
			rl: RateLimit{
				Name:   "premium",
				Limit:  20,
				Window: "30s",
				Tier:   "premium",
			},
			wantErr: false,
		},
		{
			name: "missing tier",
			rl: RateLimit{
				Name:   "free",
				Limit:  5,
				Window: "2m",
			},
			wantErr: true,
			errType: ErrRateLimitTierRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rl.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RateLimit.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != tt.errType {
				t.Errorf("RateLimit.Validate() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestRateLimitIsValid(t *testing.T) {
	validRL := RateLimit{
		Name:   "free",
		Limit:  5,
		Window: "2m",
		Tier:   "free",
	}

	if !validRL.IsValid() {
		t.Error("Expected valid RateLimit to return true from IsValid()")
	}

	invalidRL := RateLimit{
		Name:  "invalid",
		Limit: -1,
		Tier:  "free",
	}

	if invalidRL.IsValid() {
		t.Error("Expected invalid RateLimit to return false from IsValid()")
	}
}

func TestRateLimitUpdateValidate(t *testing.T) {
	tests := []struct {
		name    string
		rlu     RateLimitUpdate
		wantErr bool
		errType error
	}{
		{
			name: "valid update",
			rlu: RateLimitUpdate{
				Limit:  10,
				Window: "3m",
			},
			wantErr: false,
		},
		{
			name: "zero limit",
			rlu: RateLimitUpdate{
				Limit:  0,
				Window: "3m",
			},
			wantErr: true,
			errType: ErrRateLimitInvalid,
		},
		{
			name: "negative limit",
			rlu: RateLimitUpdate{
				Limit:  -5,
				Window: "3m",
			},
			wantErr: true,
			errType: ErrRateLimitInvalid,
		},
		{
			name: "missing window",
			rlu: RateLimitUpdate{
				Limit: 10,
			},
			wantErr: true,
			errType: ErrRateLimitWindowRequired,
		},
		{
			name: "invalid window format",
			rlu: RateLimitUpdate{
				Limit:  10,
				Window: "bad-format",
			},
			wantErr: true,
			errType: ErrRateLimitWindowInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rlu.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RateLimitUpdate.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != tt.errType {
				t.Errorf("RateLimitUpdate.Validate() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

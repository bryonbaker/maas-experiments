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
	"encoding/json"
	"fmt"
)

// TierAnnotationKey is the annotation key used to store tier information
const TierAnnotationKey = "alpha.maas.opendatahub.io/tiers"

// LLMInferenceService represents an LLMInferenceService custom resource
// @Description LLMInferenceService custom resource from KServe
type LLMInferenceService struct {
	Name      string                 `json:"name" example:"acme-dev-model"`                            // Name of the LLMInferenceService
	Namespace string                 `json:"namespace" example:"acme-inc-models"`                      // Namespace where the service is deployed
	Tiers     []string               `json:"tiers" example:"acme-dev-users-tier,acme-prod-users-tier"` // List of tiers associated with this service
	Spec      map[string]interface{} `json:"spec"`                                                     // Full spec of the LLMInferenceService
}

// ParseTiersFromAnnotation parses the tiers annotation value (JSON array string) into a slice of tier names
func ParseTiersFromAnnotation(annotationValue string) ([]string, error) {
	if annotationValue == "" {
		return []string{}, nil
	}

	var tiers []string
	if err := json.Unmarshal([]byte(annotationValue), &tiers); err != nil {
		return nil, fmt.Errorf("failed to parse tiers annotation: %w", err)
	}

	return tiers, nil
}

// HasTier checks if the service has the specified tier in its tiers list
func (l *LLMInferenceService) HasTier(tierName string) bool {
	for _, tier := range l.Tiers {
		if tier == tierName {
			return true
		}
	}
	return false
}

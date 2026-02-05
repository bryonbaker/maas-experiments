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

package storage

import (
	"context"
	"fmt"
	"maas-toolbox/internal/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// TokenRateLimitPolicy GVR (Group/Version/Resource) for dynamic client
var tokenRateLimitPolicyGVR = schema.GroupVersionResource{
	Group:    "kuadrant.io",
	Version:  "v1alpha1",
	Resource: "tokenratelimitpolicies",
}

// K8sTokenRateLimitStorage manages TokenRateLimitPolicy CRDs
type K8sTokenRateLimitStorage struct {
	dynamicClient dynamic.Interface
	Namespace     string
	PolicyName    string
}

// NewK8sTokenRateLimitStorage creates a new K8sTokenRateLimitStorage instance
func NewK8sTokenRateLimitStorage(client dynamic.Interface, namespace, policyName string) *K8sTokenRateLimitStorage {
	return &K8sTokenRateLimitStorage{
		dynamicClient: client,
		Namespace:     namespace,
		PolicyName:    policyName,
	}
}

// GetTokenRateLimitPolicy retrieves the TokenRateLimitPolicy CRD
func (k *K8sTokenRateLimitStorage) GetTokenRateLimitPolicy() (*unstructured.Unstructured, error) {
	policy, err := k.dynamicClient.Resource(tokenRateLimitPolicyGVR).
		Namespace(k.Namespace).
		Get(context.TODO(), k.PolicyName, metav1.GetOptions{})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get TokenRateLimitPolicy: %w", err)
	}
	
	return policy, nil
}

// UpdateTokenRateLimitPolicy updates the TokenRateLimitPolicy CRD
func (k *K8sTokenRateLimitStorage) UpdateTokenRateLimitPolicy(policy *unstructured.Unstructured) error {
	_, err := k.dynamicClient.Resource(tokenRateLimitPolicyGVR).
		Namespace(k.Namespace).
		Update(context.TODO(), policy, metav1.UpdateOptions{})
	
	if err != nil {
		return fmt.Errorf("failed to update TokenRateLimitPolicy: %w", err)
	}
	
	return nil
}

// GetAllTokenRateLimits retrieves all token rate limits from the policy
func (k *K8sTokenRateLimitStorage) GetAllTokenRateLimits() ([]models.TokenRateLimit, error) {
	policy, err := k.GetTokenRateLimitPolicy()
	if err != nil {
		return nil, err
	}
	
	// Navigate to spec.limits map
	limits, found, err := unstructured.NestedMap(policy.Object, "spec", "limits")
	if err != nil {
		return nil, fmt.Errorf("failed to get limits from policy: %w", err)
	}
	if !found {
		return []models.TokenRateLimit{}, nil
	}
	
	var tokenRateLimits []models.TokenRateLimit
	
	// Parse each limit entry
	for name, limitData := range limits {
		tokenRateLimit, err := k.parseTokenRateLimit(name, limitData)
		if err != nil {
			// Log error but continue parsing other limits
			continue
		}
		tokenRateLimits = append(tokenRateLimits, tokenRateLimit)
	}
	
	return tokenRateLimits, nil
}

// GetTokenRateLimit retrieves a specific token rate limit by name
func (k *K8sTokenRateLimitStorage) GetTokenRateLimit(name string) (*models.TokenRateLimit, error) {
	policy, err := k.GetTokenRateLimitPolicy()
	if err != nil {
		return nil, err
	}
	
	// Navigate to spec.limits map
	limits, found, err := unstructured.NestedMap(policy.Object, "spec", "limits")
	if err != nil {
		return nil, fmt.Errorf("failed to get limits from policy: %w", err)
	}
	if !found {
		return nil, models.ErrTokenRateLimitNotFound
	}
	
	// Get the specific limit
	limitData, found := limits[name]
	if !found {
		return nil, models.ErrTokenRateLimitNotFound
	}
	
	tokenRateLimit, err := k.parseTokenRateLimit(name, limitData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token rate limit: %w", err)
	}
	
	return &tokenRateLimit, nil
}

// CreateTokenRateLimit creates a new token rate limit in the policy
func (k *K8sTokenRateLimitStorage) CreateTokenRateLimit(trl *models.TokenRateLimit) error {
	policy, err := k.GetTokenRateLimitPolicy()
	if err != nil {
		return err
	}
	
	// Navigate to spec.limits map
	limits, found, err := unstructured.NestedMap(policy.Object, "spec", "limits")
	if err != nil {
		return fmt.Errorf("failed to get limits from policy: %w", err)
	}
	if !found {
		limits = make(map[string]interface{})
	}
	
	// Check if limit already exists
	if _, exists := limits[trl.Name]; exists {
		return models.ErrTokenRateLimitAlreadyExists
	}
	
	// Create the limit entry
	// Note: Must use int64 for numeric values in unstructured data
	limits[trl.Name] = map[string]interface{}{
		"rates": []interface{}{
			map[string]interface{}{
				"limit":  int64(trl.Limit),
				"window": trl.Window,
			},
		},
		"when": []interface{}{
			map[string]interface{}{
				"predicate": fmt.Sprintf(
					"auth.identity.tier == \"%s\" && !request.path.endsWith(\"/v1/models\")",
					trl.Tier,
				),
			},
		},
		"counters": []interface{}{
			map[string]interface{}{
				"expression": "auth.identity.userid",
			},
		},
	}
	
	// Update policy with new limits
	if err := unstructured.SetNestedMap(policy.Object, limits, "spec", "limits"); err != nil {
		return fmt.Errorf("failed to set limits in policy: %w", err)
	}
	
	// Update the policy
	return k.UpdateTokenRateLimitPolicy(policy)
}

// UpdateTokenRateLimit updates an existing token rate limit
func (k *K8sTokenRateLimitStorage) UpdateTokenRateLimit(name string, trl *models.TokenRateLimit) error {
	policy, err := k.GetTokenRateLimitPolicy()
	if err != nil {
		return err
	}
	
	// Navigate to spec.limits map
	limits, found, err := unstructured.NestedMap(policy.Object, "spec", "limits")
	if err != nil {
		return fmt.Errorf("failed to get limits from policy: %w", err)
	}
	if !found {
		return models.ErrTokenRateLimitNotFound
	}
	
	// Check if limit exists
	if _, exists := limits[name]; !exists {
		return models.ErrTokenRateLimitNotFound
	}
	
	// Update the limit entry
	// Note: Must use int64 for numeric values in unstructured data
	limits[name] = map[string]interface{}{
		"rates": []interface{}{
			map[string]interface{}{
				"limit":  int64(trl.Limit),
				"window": trl.Window,
			},
		},
		"when": []interface{}{
			map[string]interface{}{
				"predicate": fmt.Sprintf(
					"auth.identity.tier == \"%s\" && !request.path.endsWith(\"/v1/models\")",
					trl.Tier,
				),
			},
		},
		"counters": []interface{}{
			map[string]interface{}{
				"expression": "auth.identity.userid",
			},
		},
	}
	
	// Update policy with modified limits
	if err := unstructured.SetNestedMap(policy.Object, limits, "spec", "limits"); err != nil {
		return fmt.Errorf("failed to set limits in policy: %w", err)
	}
	
	// Update the policy
	return k.UpdateTokenRateLimitPolicy(policy)
}

// DeleteTokenRateLimit deletes a token rate limit from the policy
func (k *K8sTokenRateLimitStorage) DeleteTokenRateLimit(name string) error {
	policy, err := k.GetTokenRateLimitPolicy()
	if err != nil {
		return err
	}
	
	// Navigate to spec.limits map
	limits, found, err := unstructured.NestedMap(policy.Object, "spec", "limits")
	if err != nil {
		return fmt.Errorf("failed to get limits from policy: %w", err)
	}
	if !found {
		return models.ErrTokenRateLimitNotFound
	}
	
	// Check if limit exists
	if _, exists := limits[name]; !exists {
		return models.ErrTokenRateLimitNotFound
	}
	
	// Delete the limit entry
	delete(limits, name)
	
	// Update policy with modified limits
	if err := unstructured.SetNestedMap(policy.Object, limits, "spec", "limits"); err != nil {
		return fmt.Errorf("failed to set limits in policy: %w", err)
	}
	
	// Update the policy
	return k.UpdateTokenRateLimitPolicy(policy)
}

// parseTokenRateLimit parses a token rate limit from unstructured data
func (k *K8sTokenRateLimitStorage) parseTokenRateLimit(name string, limitData interface{}) (models.TokenRateLimit, error) {
	limitMap, ok := limitData.(map[string]interface{})
	if !ok {
		return models.TokenRateLimit{}, fmt.Errorf("invalid limit data format")
	}
	
	// Extract rates
	rates, _, err := unstructured.NestedSlice(limitMap, "rates")
	if err != nil || len(rates) == 0 {
		return models.TokenRateLimit{}, fmt.Errorf("failed to get rates")
	}
	
	rateMap, ok := rates[0].(map[string]interface{})
	if !ok {
		return models.TokenRateLimit{}, fmt.Errorf("invalid rate data format")
	}
	
	limit, _, err := unstructured.NestedInt64(rateMap, "limit")
	if err != nil {
		return models.TokenRateLimit{}, fmt.Errorf("failed to get limit")
	}
	
	window, _, err := unstructured.NestedString(rateMap, "window")
	if err != nil {
		return models.TokenRateLimit{}, fmt.Errorf("failed to get window")
	}
	
	// Extract tier from predicate
	whens, _, err := unstructured.NestedSlice(limitMap, "when")
	if err != nil || len(whens) == 0 {
		return models.TokenRateLimit{}, fmt.Errorf("failed to get when conditions")
	}
	
	whenMap, ok := whens[0].(map[string]interface{})
	if !ok {
		return models.TokenRateLimit{}, fmt.Errorf("invalid when data format")
	}
	
	predicate, _, err := unstructured.NestedString(whenMap, "predicate")
	if err != nil {
		return models.TokenRateLimit{}, fmt.Errorf("failed to get predicate")
	}
	
	// Extract tier from predicate string
	// Predicate format: auth.identity.tier == "free" && !request.path.endsWith("/v1/models")
	tier := extractTierFromPredicate(predicate)
	if tier == "" {
		return models.TokenRateLimit{}, fmt.Errorf("failed to extract tier from predicate")
	}
	
	return models.TokenRateLimit{
		Name:   name,
		Limit:  int(limit),
		Window: window,
		Tier:   tier,
	}, nil
}

// extractTierFromPredicate extracts the tier value from a predicate string
func extractTierFromPredicate(predicate string) string {
	// Parse: auth.identity.tier == "free" && ...
	// Simple string parsing to extract tier value between quotes
	start := -1
	for i, c := range predicate {
		if c == '"' {
			if start == -1 {
				start = i + 1
			} else {
				return predicate[start:i]
			}
		}
	}
	return ""
}

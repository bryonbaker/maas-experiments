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

// RateLimitPolicy GVR (Group/Version/Resource) for dynamic client
// NOTE: This is kuadrant.io/v1, NOT v1alpha1 like TokenRateLimitPolicy
var rateLimitPolicyGVR = schema.GroupVersionResource{
	Group:    "kuadrant.io",
	Version:  "v1",
	Resource: "ratelimitpolicies",
}

// K8sRateLimitStorage manages RateLimitPolicy CRDs
type K8sRateLimitStorage struct {
	dynamicClient dynamic.Interface
	Namespace     string
	PolicyName    string
}

// NewK8sRateLimitStorage creates a new K8sRateLimitStorage instance
func NewK8sRateLimitStorage(client dynamic.Interface, namespace, policyName string) *K8sRateLimitStorage {
	return &K8sRateLimitStorage{
		dynamicClient: client,
		Namespace:     namespace,
		PolicyName:    policyName,
	}
}

// GetRateLimitPolicy retrieves the RateLimitPolicy CRD
func (k *K8sRateLimitStorage) GetRateLimitPolicy() (*unstructured.Unstructured, error) {
	policy, err := k.dynamicClient.Resource(rateLimitPolicyGVR).
		Namespace(k.Namespace).
		Get(context.TODO(), k.PolicyName, metav1.GetOptions{})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get RateLimitPolicy: %w", err)
	}
	
	return policy, nil
}

// UpdateRateLimitPolicy updates the RateLimitPolicy CRD
func (k *K8sRateLimitStorage) UpdateRateLimitPolicy(policy *unstructured.Unstructured) error {
	_, err := k.dynamicClient.Resource(rateLimitPolicyGVR).
		Namespace(k.Namespace).
		Update(context.TODO(), policy, metav1.UpdateOptions{})
	
	if err != nil {
		return fmt.Errorf("failed to update RateLimitPolicy: %w", err)
	}
	
	return nil
}

// GetAllRateLimits retrieves all rate limits from the policy
func (k *K8sRateLimitStorage) GetAllRateLimits() ([]models.RateLimit, error) {
	policy, err := k.GetRateLimitPolicy()
	if err != nil {
		return nil, err
	}
	
	// Navigate to spec.limits map
	limits, found, err := unstructured.NestedMap(policy.Object, "spec", "limits")
	if err != nil {
		return nil, fmt.Errorf("failed to get limits from policy: %w", err)
	}
	if !found {
		return []models.RateLimit{}, nil
	}
	
	var rateLimits []models.RateLimit
	
	// Parse each limit entry
	for name, limitData := range limits {
		rateLimit, err := k.parseRateLimit(name, limitData)
		if err != nil {
			// Log error but continue parsing other limits
			continue
		}
		rateLimits = append(rateLimits, rateLimit)
	}
	
	return rateLimits, nil
}

// GetRateLimit retrieves a specific rate limit by name
func (k *K8sRateLimitStorage) GetRateLimit(name string) (*models.RateLimit, error) {
	policy, err := k.GetRateLimitPolicy()
	if err != nil {
		return nil, err
	}
	
	// Navigate to spec.limits map
	limits, found, err := unstructured.NestedMap(policy.Object, "spec", "limits")
	if err != nil {
		return nil, fmt.Errorf("failed to get limits from policy: %w", err)
	}
	if !found {
		return nil, models.ErrRateLimitNotFound
	}
	
	// Get the specific limit
	limitData, found := limits[name]
	if !found {
		return nil, models.ErrRateLimitNotFound
	}
	
	rateLimit, err := k.parseRateLimit(name, limitData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rate limit: %w", err)
	}
	
	return &rateLimit, nil
}

// CreateRateLimit creates a new rate limit in the policy
func (k *K8sRateLimitStorage) CreateRateLimit(rl *models.RateLimit) error {
	policy, err := k.GetRateLimitPolicy()
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
	if _, exists := limits[rl.Name]; exists {
		return models.ErrRateLimitAlreadyExists
	}
	
	// Create the limit entry
	// Note: Must use int64 for numeric values in unstructured data
	limits[rl.Name] = map[string]interface{}{
		"counters": []interface{}{
			map[string]interface{}{
				"expression": "auth.identity.userid",
			},
		},
		"rates": []interface{}{
			map[string]interface{}{
				"limit":  int64(rl.Limit),
				"window": rl.Window,
			},
		},
		"when": []interface{}{
			map[string]interface{}{
				// Note: Simpler predicate than TokenRateLimitPolicy (no path check)
				"predicate": fmt.Sprintf("auth.identity.tier == \"%s\"", rl.Tier),
			},
		},
	}
	
	// Update policy with new limits
	if err := unstructured.SetNestedMap(policy.Object, limits, "spec", "limits"); err != nil {
		return fmt.Errorf("failed to set limits in policy: %w", err)
	}
	
	// Update the policy
	return k.UpdateRateLimitPolicy(policy)
}

// UpdateRateLimit updates an existing rate limit
func (k *K8sRateLimitStorage) UpdateRateLimit(name string, rl *models.RateLimit) error {
	policy, err := k.GetRateLimitPolicy()
	if err != nil {
		return err
	}
	
	// Navigate to spec.limits map
	limits, found, err := unstructured.NestedMap(policy.Object, "spec", "limits")
	if err != nil {
		return fmt.Errorf("failed to get limits from policy: %w", err)
	}
	if !found {
		return models.ErrRateLimitNotFound
	}
	
	// Check if limit exists
	if _, exists := limits[name]; !exists {
		return models.ErrRateLimitNotFound
	}
	
	// Update the limit entry
	// Note: Must use int64 for numeric values in unstructured data
	limits[name] = map[string]interface{}{
		"counters": []interface{}{
			map[string]interface{}{
				"expression": "auth.identity.userid",
			},
		},
		"rates": []interface{}{
			map[string]interface{}{
				"limit":  int64(rl.Limit),
				"window": rl.Window,
			},
		},
		"when": []interface{}{
			map[string]interface{}{
				"predicate": fmt.Sprintf("auth.identity.tier == \"%s\"", rl.Tier),
			},
		},
	}
	
	// Update policy with modified limits
	if err := unstructured.SetNestedMap(policy.Object, limits, "spec", "limits"); err != nil {
		return fmt.Errorf("failed to set limits in policy: %w", err)
	}
	
	// Update the policy
	return k.UpdateRateLimitPolicy(policy)
}

// DeleteRateLimit deletes a rate limit from the policy
func (k *K8sRateLimitStorage) DeleteRateLimit(name string) error {
	policy, err := k.GetRateLimitPolicy()
	if err != nil {
		return err
	}
	
	// Navigate to spec.limits map
	limits, found, err := unstructured.NestedMap(policy.Object, "spec", "limits")
	if err != nil {
		return fmt.Errorf("failed to get limits from policy: %w", err)
	}
	if !found {
		return models.ErrRateLimitNotFound
	}
	
	// Check if limit exists
	if _, exists := limits[name]; !exists {
		return models.ErrRateLimitNotFound
	}
	
	// Delete the limit entry
	delete(limits, name)
	
	// Update policy with modified limits
	if err := unstructured.SetNestedMap(policy.Object, limits, "spec", "limits"); err != nil {
		return fmt.Errorf("failed to set limits in policy: %w", err)
	}
	
	// Update the policy
	return k.UpdateRateLimitPolicy(policy)
}

// parseRateLimit parses a rate limit from unstructured data
func (k *K8sRateLimitStorage) parseRateLimit(name string, limitData interface{}) (models.RateLimit, error) {
	limitMap, ok := limitData.(map[string]interface{})
	if !ok {
		return models.RateLimit{}, fmt.Errorf("invalid limit data format")
	}
	
	// Extract rates
	rates, _, err := unstructured.NestedSlice(limitMap, "rates")
	if err != nil || len(rates) == 0 {
		return models.RateLimit{}, fmt.Errorf("failed to get rates")
	}
	
	rateMap, ok := rates[0].(map[string]interface{})
	if !ok {
		return models.RateLimit{}, fmt.Errorf("invalid rate data format")
	}
	
	limit, _, err := unstructured.NestedInt64(rateMap, "limit")
	if err != nil {
		return models.RateLimit{}, fmt.Errorf("failed to get limit")
	}
	
	window, _, err := unstructured.NestedString(rateMap, "window")
	if err != nil {
		return models.RateLimit{}, fmt.Errorf("failed to get window")
	}
	
	// Extract tier from predicate
	whens, _, err := unstructured.NestedSlice(limitMap, "when")
	if err != nil || len(whens) == 0 {
		return models.RateLimit{}, fmt.Errorf("failed to get when conditions")
	}
	
	whenMap, ok := whens[0].(map[string]interface{})
	if !ok {
		return models.RateLimit{}, fmt.Errorf("invalid when data format")
	}
	
	predicate, _, err := unstructured.NestedString(whenMap, "predicate")
	if err != nil {
		return models.RateLimit{}, fmt.Errorf("failed to get predicate")
	}
	
	// Extract tier from predicate string
	// Predicate format: auth.identity.tier == "free"
	tier := extractTierFromPredicate(predicate)
	if tier == "" {
		return models.RateLimit{}, fmt.Errorf("failed to extract tier from predicate")
	}
	
	return models.RateLimit{
		Name:   name,
		Limit:  int(limit),
		Window: window,
		Tier:   tier,
	}, nil
}

// Note: extractTierFromPredicate is shared utility function
// It's already defined in k8s_tokenratelimit.go and works for both policies

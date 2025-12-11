package storage

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tier-to-group-admin/internal/models"

	"gopkg.in/yaml.v3"
)

// FileTierStorage implements TierStorage using a local YAML file
type FileTierStorage struct {
	FilePath string
}

// NewFileTierStorage creates a new FileTierStorage instance
func NewFileTierStorage(filePath string) *FileTierStorage {
	return &FileTierStorage{
		FilePath: filePath,
	}
}

// Load reads the tier configuration from the YAML file
func (f *FileTierStorage) Load() (*models.TierConfig, error) {
	data, err := os.ReadFile(f.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &models.TierConfig{Tiers: []models.Tier{}}, nil
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the YAML structure that matches ConfigMap format
	// The ConfigMap has a "data" field with "tiers" as a YAML string
	var configMap struct {
		Data struct {
			Tiers string `yaml:"tiers"`
		} `yaml:"data"`
	}

	if err := yaml.Unmarshal(data, &configMap); err != nil {
		// Try parsing as direct TierConfig (for simpler format)
		var directConfig models.TierConfig
		if err2 := yaml.Unmarshal(data, &directConfig); err2 != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
		return &directConfig, nil
	}

	// Parse the tiers YAML string
	// The tiers field contains a YAML array, not a TierConfig struct
	var tiers []models.Tier
	if err := yaml.Unmarshal([]byte(configMap.Data.Tiers), &tiers); err != nil {
		// If tiers is empty or "[]", return empty config
		if configMap.Data.Tiers == "" || configMap.Data.Tiers == "[]" {
			return &models.TierConfig{Tiers: []models.Tier{}}, nil
		}
		return nil, fmt.Errorf("failed to parse tiers YAML: %w", err)
	}

	return &models.TierConfig{Tiers: tiers}, nil
}

// Save writes the tier configuration to the YAML file
func (f *FileTierStorage) Save(config *models.TierConfig) error {
	// Format as ConfigMap structure
	configMap := struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name      string `yaml:"name"`
			Namespace string `yaml:"namespace"`
		} `yaml:"metadata"`
		Data struct {
			Tiers string `yaml:"tiers"`
		} `yaml:"data"`
	}{
		APIVersion: "v1",
		Kind:       "ConfigMap",
		Metadata: struct {
			Name      string `yaml:"name"`
			Namespace string `yaml:"namespace"`
		}{
			Name:      "tier-to-group-mapping",
			Namespace: "maas-api",
		},
	}

	// Marshal tiers to YAML string with 2-space indentation
	var tiersBuffer bytes.Buffer
	tiersEncoder := yaml.NewEncoder(&tiersBuffer)
	tiersEncoder.SetIndent(2)
	if err := tiersEncoder.Encode(config.Tiers); err != nil {
		return fmt.Errorf("failed to marshal tiers: %w", err)
	}
	tiersEncoder.Close()
	// Remove document separator and trailing newline if present
	tiersYAML := tiersBuffer.String()
	tiersYAML = strings.TrimPrefix(tiersYAML, "---\n")
	tiersYAML = strings.TrimSuffix(tiersYAML, "\n")
	configMap.Data.Tiers = tiersYAML

	// Marshal entire config map with 2-space indentation
	var dataBuffer bytes.Buffer
	dataEncoder := yaml.NewEncoder(&dataBuffer)
	dataEncoder.SetIndent(2)
	if err := dataEncoder.Encode(&configMap); err != nil {
		return fmt.Errorf("failed to marshal config map: %w", err)
	}
	dataEncoder.Close()
	data := dataBuffer.Bytes()

	// Ensure the directory exists before writing
	dir := filepath.Dir(f.FilePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write to file (creates file if it doesn't exist)
	if err := os.WriteFile(f.FilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}


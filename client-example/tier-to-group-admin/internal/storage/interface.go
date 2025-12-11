package storage

import "tier-to-group-admin/internal/models"

// TierStorage defines the interface for tier persistence
// This abstraction allows swapping between file-based and Kubernetes storage
type TierStorage interface {
	// Load retrieves the tier configuration from storage
	Load() (*models.TierConfig, error)

	// Save persists the tier configuration to storage
	Save(config *models.TierConfig) error
}


package anime_processor

import "github.com/Flagsmith/flagsmith-go-client/v2"

//go:generate mockgen -source=flagsmith_client_interface.go -destination=mocks/mock_flagsmith_client.go -package=mocks

// FlagSmithClient defines the interface for Flagsmith client operations
type FlagSmithClient interface {
	GetEnvironmentFlags() (flagsmith.Flags, error)
}

// Flags defines the interface for Flagsmith flags operations  
type Flags interface {
	IsFeatureEnabled(featureName string) (bool, error)
}
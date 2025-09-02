package anime_processor

import "github.com/Flagsmith/flagsmith-go-client/v2"

//go:generate mockgen -source=flagsmith_interfaces.go -destination=mocks/mock_flagsmith.go -package=mocks

// FlagsmithClientInterface defines the interface we need from Flagsmith client
type FlagsmithClientInterface interface {
	GetEnvironmentFlags() (flagsmith.Flags, error)
}

// FlagsInterface defines the interface we need from Flagsmith flags
type FlagsInterface interface {
	IsFeatureEnabled(featureName string) (bool, error)
}
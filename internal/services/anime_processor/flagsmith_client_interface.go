package anime_processor

//go:generate mockgen -source=flagsmith_client_interface.go -destination=mocks/mock_flagsmith_client.go -package=mocks

// FlagSmithClient defines the interface for Flagsmith client operations
type FlagSmithClient interface {
	GetEnvironmentFlags() (Flags, error)
}

// Flags defines the interface for Flagsmith flags operations
type Flags interface {
	IsFeatureEnabled(featureName string) (bool, error)
}

// FlagsmithClientWrapper wraps a real Flagsmith client to implement our interface
type FlagsmithClientWrapper struct {
	Client interface {
		GetEnvironmentFlags() (interface{ IsFeatureEnabled(string) (bool, error) }, error)
	}
}

// GetEnvironmentFlags wraps the real client's method
func (w *FlagsmithClientWrapper) GetEnvironmentFlags() (Flags, error) {
	flags, err := w.Client.GetEnvironmentFlags()
	if err != nil {
		return nil, err
	}
	return flags, nil
}
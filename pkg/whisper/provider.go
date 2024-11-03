package whisper

import (
	"context"
	"sync"
)

type WhisperHandler interface {
	HandleWhisper(ctx context.Context, input WhisperInput) (*WhisperOutput, error)
}

var DefaultWhisperProvider = &WhisperProvider{}

type WhisperProvider struct {
	providerMap sync.Map
}

func (provider *WhisperProvider) Register(name string, f WhisperHandler) {
	provider.providerMap.Store(name, f)
}

func (provider *WhisperProvider) Get(name string) (WhisperHandler, bool) {
	f, ok := provider.providerMap.Load(name)
	if !ok {
		return nil, ok
	}

	return f.(WhisperHandler), ok
}

func (provider *WhisperProvider) AllProviderNames() []string {
	var names []string

	provider.providerMap.Range(func(key, value any) bool {
		names = append(names, key.(string))

		return true
	})

	return names
}

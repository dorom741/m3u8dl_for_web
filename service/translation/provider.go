package translation

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var _ ITranslation = &TranslationProviderHub{}

type TranslationProviderHub struct {
	providerList []ITranslation
	mu           sync.RWMutex
}

func NewTranslationProviderHub(config TranslationProviderHubConfig, httpClient *http.Client) (*TranslationProviderHub, error) {
	hub := &TranslationProviderHub{
		// mu:           sync.RWMutex{},
		providerList: make([]ITranslation, 0),
	}

	for _, p := range config.Providers {
		if !p.Enable {
			continue
		}

		switch strings.ToLower(p.Type) {
		case "deeplx":
			if p.DeepLX == nil {
				return nil, fmt.Errorf("provider %s: missing deepLX config", p.Name)
			}
			t := NewDeepLXTranslation(p.DeepLX, httpClient)
			logrus.Infof("register translation provider '%s' type '%s'", p.Name, p.Type)
			hub.RegisterTranslationProvider(t)

		case "openai", "openaicompatible", "openai-compatible":
			if p.OpenAiCompatible == nil {
				return nil, fmt.Errorf("provider %s: missing openAiCompatible config", p.Name)
			}
			t := NewOpenAiCompatibleTranslation(p.OpenAiCompatible)
			logrus.Infof("register translation provider '%s' type '%s'", p.Name, p.Type)
			hub.RegisterTranslationProvider(t)

		default:
			return nil, fmt.Errorf("unknown provider type: %s", p.Type)
		}
	}

	return hub, nil
}

func (hub *TranslationProviderHub) SupportMultipleTextBySeparator() (bool, string) {
	return false, ""
}

func (hub *TranslationProviderHub) RegisterTranslationProvider(provider ITranslation) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	hub.providerList = append(hub.providerList, provider)
}

func (hub *TranslationProviderHub) GetRandomProvider() ITranslation {
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	providerListLen := len(hub.providerList)

	if providerListLen == 0 {
		return nil
	}

	index := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(providerListLen)
	return hub.providerList[index]
}

func (hub *TranslationProviderHub) Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	provider := hub.GetRandomProvider()
	if provider == nil {
		return "", fmt.Errorf("none provider registered")
	}

	var (
		lastErr        error
		maxAttempts    = 3
		attempts       = 0
		translatedText = ""
	)
	for attempts < maxAttempts {
		attempts++
		translatedText, lastErr = provider.Translate(ctx, text, sourceLang, targetLang)
		if lastErr == nil {
			return translatedText, nil
		}
	}

	return "", fmt.Errorf("translate failed after %d attempts of text '%s',last error: %v", attempts, text, lastErr)

}

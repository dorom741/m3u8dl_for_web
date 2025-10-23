package translation

import (
	"context"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// ITranslation 是翻译实现必须实现的接口
type ITranslation interface {
	GetName() string
	Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error)
	SupportMultipleTextBySeparator() (bool, string)
}

type TranslationProviderConfig struct {
	Name                    string                   `yaml:"name"`
	Enable                  bool                     `yaml:"enable"`
	Type                    string                   `yaml:"type"` // deeplx, openai,google
	DeepLX                  *DeepLXConfig            `yaml:"deepLXConfig,omitempty"`
	OpenAiCompatible        *OpenAiCompatibleConfig  `yaml:"openaiConfig,omitempty"`
	GoogleTranslationConfig *GoogleTranslationConfig `yaml:"googleTranslationConfig,omitempty"`
}

type TranslationProviderHubConfig struct {
	Providers []TranslationProviderConfig `yaml:"providers"`
}

type DeepLXResult struct {
	Data string
}

type DeepLXConfig struct {
	UrlsFile              string   `yaml:"urlsFile"`
	Urls                  []string `yaml:"urls"`
	ApiKey                string   `yaml:"apiKey"`
	RPM                   int      `yaml:"RPM"`
	MultipleTextSeparator string   `yaml:"multipleTextSeparator"`
}

func (deepLXConfig *DeepLXConfig) ParseUrlFile() {
	if len(deepLXConfig.UrlsFile) == 0 {
		return
	}

	data, err := os.ReadFile(deepLXConfig.UrlsFile)
	if err != nil {
		logrus.Warnf("parse DeepLX urls file error:%s", err)
		return
	}

	urls := strings.Split(string(data), "\n")
	deepLXConfig.Urls = append(urls, deepLXConfig.Urls...)
	deepLXConfig.Urls = removeEmptyStrings(deepLXConfig.Urls)
}

func (deepLXConfig *DeepLXConfig) WriteUrlFile(urls []string) {
	if len(deepLXConfig.UrlsFile) == 0 {
		return
	}

	err := os.WriteFile(deepLXConfig.UrlsFile, []byte(strings.Join(urls, "\n")), os.ModePerm)
	if err != nil {
		logrus.Warnf("write DeepLX urls file error:%s", err)
	}

}

type OpenAiCompatibleConfig struct {
	BaseUrl               string `yaml:"baseUrl"`
	ApiKey                string `yaml:"apiKey"`
	Model                 string `yaml:"model"`
	SystemPrompt          string `yaml:"systemPrompt"`
	Prompt                string `yaml:"prompt"`
	ContextLen            int    `yaml:"contextLen"`
	RPM                   int    `yaml:"RPM"`
	MultipleTextSeparator string `yaml:"multipleTextSeparator"`
}

func removeEmptyStrings(strs []string) []string {
	var result []string
	for _, s := range strs {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, s)
		}
	}
	return result
}

type GoogleTranslationConfig struct {
	RPM                   int    `yaml:"RPM"`
	MultipleTextSeparator string `yaml:"multipleTextSeparator"`
}

package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var _ ITranslation = &GoogleTranslation{}

type GoogleTranslation struct {
	baseUrl string
	client  *http.Client
	limiter *rate.Limiter
	config  *GoogleTranslationConfig
}

func NewGoogleTranslation(config *GoogleTranslationConfig, httpClient *http.Client) *GoogleTranslation {
	translation := &GoogleTranslation{
		config:  config,
		baseUrl: "https://translate.googleapis.com/translate_a/single?client=gtx&dt=t&dt=bd&dt=md&dt=ex&sl=%s&tl=%s&q=%s",
		client:  httpClient,
	}

	if translation.client == nil {
		translation.client = http.DefaultClient
	}

	if config.RPM > 0 {
		burst := (config.RPM + 60 - 1) / 60
		translation.limiter = rate.NewLimiter(rate.Every(time.Second*time.Duration(int(60/config.RPM)+1)), burst)
		logrus.Infof("enable google translation rate limit RPM %d burst %d", config.RPM, burst)

	}

	return translation
}

func (translation *GoogleTranslation) GetName() string {
	return "GoogleTranslate"
}

func (translation *GoogleTranslation) httpClientDo(req *http.Request) (*http.Response, error) {
	if translation.limiter != nil {
		if err := translation.limiter.Wait(req.Context()); err != nil {
			return nil, err
		}
	}
	return translation.client.Do(req)
}

func (translation *GoogleTranslation) SupportMultipleTextBySeparator() (bool, string) {
	return len(translation.config.MultipleTextSeparator) > 0, translation.config.MultipleTextSeparator
}

func (translation *GoogleTranslation) translateOnce(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	var (
		urlStr = fmt.Sprintf(translation.baseUrl, sourceLang, targetLang, url.QueryEscape(text))
		result []any
	)

	println(urlStr)
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("google translation: create request for %s failed: %v", urlStr, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := translation.httpClientDo(req)
	if err != nil {
		return "", fmt.Errorf("google translation: request to %s failed: %v", urlStr, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google translation: status %d,body %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result) <= 0 {
		return "", fmt.Errorf("google translation result format error:%s", string(body))
	}

	translationText := ""
	for _, line := range result[0].([]interface{}) {
		translatedLine := line.([]interface{})[0]
		translationText += translatedLine.(string)
	}

	return translationText, nil
}

func (translation *GoogleTranslation) Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	var (
		lastErr     error
		attempts    = 0
		maxAttempts = 3
	)

	if sourceLang == "" {
		sourceLang = "auto"
	}

	for attempts < maxAttempts {
		attempts++

		res, err := translation.translateOnce(ctx, text, sourceLang, targetLang)
		if err == nil {
			return res, nil
		}

		lastErr = err
	}

	return "", fmt.Errorf("google translation: failed after %d attempts: last error: %v", attempts, lastErr)
}

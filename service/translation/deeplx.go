package translation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var _ ITranslation = &DeepLXTranslation{}

type DeepLXTranslation struct {
	urls    []string
	mu      sync.Mutex
	client  *http.Client
	limiter *rate.Limiter
	config  *DeepLXConfig
}

func NewDeepLXTranslation(config *DeepLXConfig, httpClient *http.Client) *DeepLXTranslation {
	config.ParseUrlFile()
	logrus.Debugf("DeepLX all url:%+v", config.Urls)
	translation := &DeepLXTranslation{
		config: config,
		urls:   config.Urls,
		client: httpClient,
	}

	if translation.client == nil {
		translation.client = http.DefaultClient
	}

	if config.RPM > 0 {
		burst := (config.RPM + 60 - 1) / 60
		translation.limiter = rate.NewLimiter(rate.Every(time.Second*time.Duration(int(60/config.RPM)+1)), burst)
		logrus.Infof("enable DeepLX translation rate limit RPM %d burst %d", config.RPM, burst)

	}

	return translation
}

type Result struct {
	Data string
}

func (translation *DeepLXTranslation) httpClientDo(req *http.Request) (*http.Response, error) {
	if translation.limiter != nil {
		if err := translation.limiter.Wait(req.Context()); err != nil {
			return nil, err
		}
	}
	return translation.client.Do(req)
}

func (translation *DeepLXTranslation) getRandomUrl() string {
	translation.mu.Lock()
	defer translation.mu.Unlock()

	if len(translation.urls) == 0 {
		return ""
	}

	index := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(translation.urls))
	return translation.urls[index]
}

func (translation *DeepLXTranslation) removeUrl(url string) {
	translation.mu.Lock()
	defer translation.mu.Unlock()

	for i, p := range translation.urls {
		if p == url {
			translation.urls = append(translation.urls[:i], translation.urls[i+1:]...)
			break
		}
	}

	translation.config.WriteUrlFile(translation.urls)
}

func (translation *DeepLXTranslation) SupportMultipleTextBySeparator() (bool, string) {
	return false, "\n"
}

func (translation *DeepLXTranslation) translateOnce(ctx context.Context, url string, postData map[string]string) (string, error) {
	postDataReader := new(bytes.Buffer)
	if err := json.NewEncoder(postDataReader).Encode(postData); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, postDataReader)
	if err != nil {
		return "", fmt.Errorf("DeepLX: create request for %s failed: %v", url, err)
	}

	resp, err := translation.httpClientDo(req)
	if err != nil {
		return "", fmt.Errorf("DeepLX: request to %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepLX: status %d", resp.StatusCode)
	}

	result := new(Result)
	if err := json.Unmarshal(body, result); err != nil {
		return "", err
	}

	if strings.TrimSpace(result.Data) == "" {
		return "", fmt.Errorf("DeepLX: empty result")
	}

	return result.Data, nil
}

func (translation *DeepLXTranslation) Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	postData := map[string]string{
		"text":        text,
		"source_lang": sourceLang,
		"target_lang": targetLang,
	}

	translation.mu.Lock()
	numUrls := len(translation.urls)
	translation.mu.Unlock()

	if numUrls == 0 {
		return "", fmt.Errorf("no DeepLX urls configured")
	}

	var (
		lastErr  error
		attempts = 0
	)

	if numUrls == 1 {
		res, err := translation.translateOnce(ctx, translation.getRandomUrl(), postData)
		if err != nil {
			return "", fmt.Errorf("DeepLX translate error: %v", err)
		}
		return res, nil
	}

	for attempts < numUrls {
		attempts++

		url := translation.getRandomUrl()
		if url == "" {
			break
		}

		res, err := translation.translateOnce(ctx, url, postData)
		if err == nil {
			return res, nil
		}

		translation.removeUrl(url)
		lastErr = err
	}

	return "", fmt.Errorf("all DeepLX urls failed after %d attempts: last error: %v", attempts, lastErr)
}

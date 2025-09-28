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

	"m3u8dl_for_web/conf"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var _ ITranslation = &DeepLXTranslation{}

type DeepLXTranslation struct {
	urls    []string
	mu      sync.Mutex
	client  *http.Client
	limiter *rate.Limiter
	config  *conf.DeepLXConfig
}

func NewDeepLXTranslation(config *conf.DeepLXConfig, httpClient *http.Client) *DeepLXTranslation {
	config.ParseUrlFile()
	logrus.Debugf("all DeepLX url:%+v", config.Urls)
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

func (translation *DeepLXTranslation) Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	postData := map[string]string{
		"text":        text,
		"source_lang": sourceLang,
		"target_lang": targetLang,
	}

	// determine number of attempts = initial number of urls
	translation.mu.Lock()
	maxAttempts := len(translation.urls)
	translation.mu.Unlock()

	if maxAttempts == 0 {
		return "", fmt.Errorf("no DeepLX urls configured")
	}

	var lastErr error
	attempts := 0

	for attempts < maxAttempts {
		attempts++

		url := translation.getRandomUrl()
		if url == "" {
			break
		}

		// prepare body for this attempt
		postDataReader := new(bytes.Buffer)
		if err := json.NewEncoder(postDataReader).Encode(postData); err != nil {
			return "", err
		}

		req, err := http.NewRequestWithContext(ctx,"POST", url, postDataReader)
		if err != nil {
			logrus.Warnf("DeepLX: create request for %s failed: %v", url, err)
			return "", err
		}

		resp, err := translation.httpClientDo(req)
		if err != nil {
			logrus.Warnf("DeepLX: request to %s failed: %v", url, err)
			translation.removeUrl(url)
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logrus.Warnf("DeepLX: read body from %s failed: %v", url, err)
			translation.removeUrl(url)
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			logrus.Warnf("DeepLX: request to %s error status:%d body:%s", url, resp.StatusCode, string(body))
			translation.removeUrl(url)
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}

		result := new(Result)
		if err := json.Unmarshal(body, result); err != nil {
			logrus.Warnf("DeepLX: parse response from %s failed: %v", url, err)
			translation.removeUrl(url)
			lastErr = err
			continue
		}

		if strings.TrimSpace(result.Data) == "" {
			logrus.Warnf("DeepLX: empty result from %s", url)
			translation.removeUrl(url)
			lastErr = fmt.Errorf("empty result")
			continue
		}

		return result.Data, nil
	}

	if lastErr != nil {
		return "", fmt.Errorf("all DeepLX urls failed after %d attempts: last error: %v", attempts, lastErr)
	}
	return "", fmt.Errorf("all DeepLX urls failed after %d attempts", attempts)
}

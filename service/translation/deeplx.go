package translation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"m3u8dl_for_web/conf"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var _ ITranslation = &DeepLXTranslation{}

type DeepLXTranslation struct {
	Url     string
	client  *http.Client
	limiter *rate.Limiter
}

func NewDeepLXTranslation(config *conf.DeepLXConfig, httpClient *http.Client) *DeepLXTranslation {
	translation := &DeepLXTranslation{
		Url:    config.Url,
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

func (translation *DeepLXTranslation) SupportMultipleTextBySeparator() (bool, string) {
	return false, "\n"
}

func (translation *DeepLXTranslation) Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	var (
		err            error
		postDataReader = new(bytes.Buffer)
		postData       = map[string]string{
			"text":        text,
			"source_lang": sourceLang,
			"target_lang": targetLang,
		}
		result = new(Result)
	)

	if err = json.NewEncoder(postDataReader).Encode(postData); err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", translation.Url, postDataReader)
	if err != nil {
		return "", err
	}
	request.WithContext(ctx)

	response, err := translation.httpClientDo(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request error status:%d err:%s", response.StatusCode, body)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return "", err
	}

	return result.Data, nil
}

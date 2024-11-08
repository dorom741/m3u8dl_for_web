package translation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var _ ITranslation = &DeepLXTranslation{}

type DeepLXTranslation struct {
	Url    string
	client *http.Client
}

func NewDeepLXTranslation(deeplXUrl string, httpClient *http.Client) *DeepLXTranslation {
	translation := &DeepLXTranslation{
		Url:    deeplXUrl,
		client: httpClient,
	}

	if translation.client == nil {
		translation.client = http.DefaultClient
	}

	return translation
}

type Result struct {
	Data string
}

func (translation *DeepLXTranslation) SupportMultipleTextByPunctuation() (bool, string) {
	return true, "\n"
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

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		response, err := translation.client.Do(request)
		if err != nil {
			return "", err
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return "", err
		}

		if response.StatusCode == http.StatusServiceUnavailable {
			time.Sleep(1 * time.Second)
			continue
		}

		if response.StatusCode != http.StatusOK {
			return "", fmt.Errorf("request error status:%d err:%s", response.StatusCode, body)
		}

		if err := json.Unmarshal(body, result); err != nil {
			return "", err
		}

	}

	return result.Data, nil
}

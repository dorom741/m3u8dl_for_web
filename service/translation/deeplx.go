package translation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (translation *DeepLXTranslation) Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	var (
		postDataReader = new(bytes.Buffer)
		postData       = map[string]string{
			"text":        text,
			"source_lang": sourceLang,
			"target_lang": targetLang,
		}
	)
	_ = json.NewEncoder(postDataReader).Encode(postData)

	request, err := http.NewRequest("POST", translation.Url, postDataReader)
	if err != nil {
		return "", err
	}
	request.WithContext(ctx)

	response, err := translation.client.Do(request)
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

	result := new(Result)

	if err := json.Unmarshal(body, result); err != nil {
		return "", err
	}

	return result.Data, nil
}

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"m3u8dl_for_web/infra"

	"github.com/conneroisu/groq-go"
)

type GroqService struct {
	apiKey    string
	cachePath string
	client    *groq.Client
}

func NewGroqService(apiKey string, cachePath string, proxyURLString string) (*GroqService, error) {
	transport := &http.Transport{}

	proxyURL, err := url.Parse(proxyURLString)
	if err == nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	client, err := groq.NewClient(apiKey, groq.WithClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &GroqService{
		apiKey:    apiKey,
		client:    client,
		cachePath: cachePath,
	}, nil
}

func (groqService *GroqService) AudioTranslation(ctx context.Context, audioPath string) (*groq.AudioResponse, error) {
	var resp *groq.AudioResponse
	if err := groqService.readCache(audioPath, &resp); err != nil {
		return nil, err
	} else if resp != nil {
		return resp, nil
	}

	response, err := groqService.client.CreateTranslation(ctx, groq.AudioRequest{
		Model:    groq.ModerationTextStable,
		FilePath: audioPath,
		Format:   groq.AudioResponseFormatVerboseJSON,
		// Prompt:   "english and mandarin",
	})
	if err != nil {
		return nil, err
	}
	infra.Logger.Infof("response %+v", response)
	if err := groqService.writeCache(response, audioPath); err != nil {
		return nil, err
	}

	return &response, nil
}

func (groqService *GroqService) writeCache(data interface{}, originalFilepath string) error {
	filename := filepath.Base(originalFilepath) + "_groqcache.json"
	cacheFilePath := path.Join(groqService.cachePath, filename)
	cacheFile, err := os.OpenFile(cacheFilePath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer cacheFile.Close()
	return json.NewEncoder(cacheFile).Encode(data)
}

func (groqService *GroqService) readCache(originalFilepath string, v any) error {
	filename := filepath.Base(originalFilepath) + "_groqcache.json"
	cacheFilePath := path.Join(groqService.cachePath, filename)
	cacheFile, err := os.Open(cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer cacheFile.Close()

	return json.NewDecoder(cacheFile).Decode(v)
}

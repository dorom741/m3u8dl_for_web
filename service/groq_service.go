package service

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/pkg/whisper"

	"github.com/conneroisu/groq-go"
)

var _ whisper.WhisperHandler = &GroqService{}

type GroqService struct {
	apiKey string
	cache  *infra.FileCache
	client *groq.Client
}

func NewGroqService(apiKey string, cache *infra.FileCache, proxyURLString string) (*GroqService, error) {
	transport := &http.Transport{}

	if len(proxyURLString) > 0 && strings.HasPrefix(proxyURLString, "http") {
		if proxyURL, err := url.Parse(proxyURLString); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	client, err := groq.NewClient(apiKey, groq.WithClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &GroqService{
		apiKey: apiKey,
		client: client,
		cache:  cache,
	}, nil
}

func (service *GroqService) cacheKey(input whisper.WhisperInput) string {
	prefix := "groqcache"

	filename := filepath.Base(input.FilePath)
	ext := filepath.Ext(input.FilePath)

	return prefix + "_" + filename[:len(filename)-len(ext)]
}

// 最大 25MB
func (service *GroqService) MaximumFileSize() int64 {
	// 1024 * 1024 * 25
	return 26214400
}

func (service *GroqService) HandleWhisper(ctx context.Context, input whisper.WhisperInput) (*whisper.WhisperOutput, error) {
	var (
		resp     *groq.AudioResponse
		cacheKey = service.cacheKey(input)
	)

	if err := service.cache.Get(cacheKey, &resp); err != nil {
		return nil, err
	} else if resp != nil {
		logrus.Infof("handle whisper using cache: %s", cacheKey)
		return service.GetWhisperOutput(*resp), nil
	}

	response, err := service.client.CreateTranscription(ctx, groq.AudioRequest{
		Model:       groq.WhisperLargeV3,
		Format:      groq.AudioResponseFormatVerboseJSON,
		FilePath:    input.FilePath,
		Language:    input.Language,
		Temperature: input.Temperature,
		Prompt:      input.Prompt,
	})
	logrus.Infof("response %+v", response)

	if err != nil {
		return nil, err
	}
	if err := service.cache.Set(cacheKey, response); err != nil {
		return nil, err
	}

	return service.GetWhisperOutput(response), nil
}

func (service *GroqService) GetWhisperOutput(response groq.AudioResponse) *whisper.WhisperOutput {
	segmentList := make([]whisper.Segment, 0, len(response.Segments))

	for i, segment := range response.Segments {
		segmentList = append(segmentList, whisper.Segment{
			Num:   i,
			Start: segment.Start,
			End:   segment.End,
			Text:  segment.Text,
		})
	}

	return &whisper.WhisperOutput{Segments: segmentList, Duration: response.Duration}
}

package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/pkg/whisper"

	"github.com/conneroisu/groq-go"
)

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

	if input.Reader != nil {
		h := sha256.New()

		// 读取数据并写入哈希对象
		if _, err := io.Copy(h, input.Reader); err != nil {
			return ""
		}
		return fmt.Sprintf("%s_%x", prefix, h.Sum(nil))
	}

	return prefix + "_" + input.FilePath[:len(input.FilePath)-len(filepath.Ext(input.FilePath))]
}

func (service *GroqService) HandleWhisper(ctx context.Context, input whisper.WhisperInput) (whisper.WhisperOutput, error) {
	var (
		resp     *groq.AudioResponse
		cacheKey = service.cacheKey(input)
	)

	if err := service.cache.Get(cacheKey, &resp); err != nil {
		return nil, err
	} else if resp != nil {
		// fmt.Printf("audio file '%s'  translation  use cache: %+v\n", audioPath, resp)
		return &GroqWhisperOutput{AudioResponse: resp}, nil
	}

	response, err := service.client.CreateTranscription(ctx, groq.AudioRequest{
		Model:       groq.WhisperLargeV3,
		Format:      groq.AudioResponseFormatVerboseJSON,
		FilePath:    input.FilePath,
		Language:    input.Language,
		Temperature: input.Temperature,
		Prompt:      input.Prompt,
	})
	if err != nil {
		return nil, err
	}
	infra.Logger.Infof("response %+v", response)
	if err := service.cache.Set(cacheKey, response); err != nil {
		return nil, err
	}

	return &GroqWhisperOutput{AudioResponse: &response}, nil
}

// func (service *GroqService) writeCache(data interface{}, originalFilepath string) error {
// 	filename := filepath.Base(originalFilepath) + "_groqcache.json"
// 	cacheFilePath := path.Join(service.cachePath, filename)
// 	cacheFile, err := os.OpenFile(cacheFilePath, os.O_CREATE|os.O_RDWR, os.ModePerm)
// 	if err != nil {
// 		return err
// 	}
// 	defer cacheFile.Close()
// 	return json.NewEncoder(cacheFile).Encode(data)
// }

// func (groqService *GroqService) readCache(originalFilepath string, v any) error {
// 	filename := filepath.Base(originalFilepath) + "_groqcache.json"
// 	cacheFilePath := path.Join(groqService.cachePath, filename)
// 	cacheFile, err := os.Open(cacheFilePath)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return nil
// 		}
// 		return err
// 	}
// 	defer cacheFile.Close()

// 	return json.NewDecoder(cacheFile).Decode(v)
// }

type GroqWhisperOutput struct {
	*groq.AudioResponse
}

func (output *GroqWhisperOutput) GetSegmentList() whisper.Segments {
	segmentList := make(whisper.Segments, 0, len(output.Segments))

	for i, segment := range output.Segments {
		segmentList = append(segmentList, whisper.Segment{
			Num:   i,
			Start: segment.Start,
			End:   segment.End,
			Text:  segment.Text,
		})
	}

	return segmentList
}

func (output *GroqWhisperOutput) GetDuration() float64 {
	return output.Duration
}

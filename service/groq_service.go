package service

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

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
		response *groq.AudioResponse
		cacheKey = service.cacheKey(input)
	)

	if err := service.cache.Get(cacheKey, &response); err != nil {
		return nil, err
	} else if response != nil {
		logrus.Infof("handle whisper using cache: %s", cacheKey)
		return service.GetWhisperOutput(*response), nil
	}

	for i := 0; i < 3; i++ {
		tempResponse, err := service.client.CreateTranscription(ctx, groq.AudioRequest{
			Model:       groq.WhisperLargeV3,
			Format:      groq.AudioResponseFormatVerboseJSON,
			FilePath:    input.FilePath,
			Language:    input.Language,
			Temperature: input.Temperature,
			Prompt:      input.Prompt,
		})
		if err == nil {
			response = &tempResponse
			break
		}

		//error example:
		//status code: 429, message: Rate limit reached for model `whisper-large-v3` in
		//organization `org_01j9znf32dft8bty4veb2z96pr` on seconds of audio per day (ASPD):
		//Limit 28800, Used 28598, Requested 818. Please try again in 30m46.361999999s.
		//Visit https://console.groq.com/docs/rate-limits for more information.
		if !strings.Contains(err.Error(), "429") {
			return nil, err
		}

		waitDuration := service.parseDuration(err.Error())
		logrus.Warnf("groq service rate limit reached,wait for %s retry", waitDuration)
		time.Sleep(waitDuration)
	}

	if err := service.cache.Set(cacheKey, response); err != nil {
		return nil, err
	}

	return service.GetWhisperOutput(*response), nil
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

func (service *GroqService) parseDuration(errString string) time.Duration {
	toFloat := func(s string) float64 {
		float, _ := strconv.ParseFloat(s, 10)
		return float
	}

	// 正则表达式，用于匹配时间格式
	re := regexp.MustCompile(`(\d+)([smh])`)

	// 匹配结果
	matches := re.FindAllStringSubmatch(errString, -1)

	// 初始化时长
	var totalDuration time.Duration

	// 解析匹配结果
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		// 获取数值和单位
		value := match[1]
		unit := match[2]

		// 将值转换为整数
		var duration time.Duration
		switch unit {
		case "s":
			duration = time.Duration(toFloat(value)) * time.Second
		case "m":
			duration = time.Duration(toFloat(value)) * time.Minute
		case "h":
			duration = time.Duration(toFloat(value)) * time.Hour
		}

		// 累加总时长
		totalDuration += duration
	}

	return totalDuration

}

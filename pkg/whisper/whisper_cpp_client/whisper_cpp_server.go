package whispercppclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"m3u8dl_for_web/pkg/whisper"
	"mime/multipart"
	"net/http"

	"github.com/sirupsen/logrus"
)

var _ whisper.WhisperHandler = &WhisperCppClient{}

type WhisperCppClient struct {
	baseUrl       string
	client        *http.Client
	splitDuration int64
	debugMode     bool
}

func NewWhisperCppClient(config *WhisperCppClientConfig, httpClient *http.Client) *WhisperCppClient {
	cppClient := &WhisperCppClient{
		baseUrl:       config.BaseUrl,
		splitDuration: config.SplitDuration,
		debugMode:     config.DebugMode,
		client:        httpClient,
	}

	if cppClient.client == nil {
		cppClient.client = http.DefaultClient
	}

	return cppClient
}

// 8/16/16000 * splitDuration
func (whisperCppClient *WhisperCppClient) MaximumFileSize() int64 {
	return 32000 * whisperCppClient.splitDuration
}

func (whisperCppClient *WhisperCppClient) HandleWhisper(ctx context.Context, input whisper.WhisperInput) (*whisper.WhisperOutput, error) {
	var (
		pr, pw = io.Pipe()
		writer = multipart.NewWriter(pw)
		req    = WhisperRequest{
			DebugMode:   whisperCppClient.debugMode,
			Prompt:      input.Prompt,
			Language:    input.Language,
			Temperature: float64(input.Temperature),
			SplitOnWord: true,
			MaxLen:      25,
			PrintProgress: true,
			// PrintRealtime:  true,
			// PrintSpecial:   true,
			ResponseFormat: "verbose_json",
		}
		whisperResponse = &WhisperResponse{}
	)

	logrus.Debugf("whisperCppClient req:%+v", req)

	go func() {
		_ = req.WriteMultipartWithFile(writer, "file", input.FilePath)
		writer.Close()
		pw.Close()
	}()

	httpReq, err := http.NewRequest("POST", whisperCppClient.baseUrl, pr)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := whisperCppClient.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error status:%d err:%s", response.StatusCode, body)
	}

	if err := json.Unmarshal(body, whisperResponse); err != nil {
		return nil, err
	}

	var segments = make([]whisper.Segment, len(whisperResponse.Segments))

	for i, segment := range whisperResponse.Segments {
		segments[i] = whisper.Segment{
			Num:   int(segment.ID),
			Start: segment.Start,
			End:   segment.End,
			Text:  segment.Text,
		}

	}

	return &whisper.WhisperOutput{Segments: segments, Duration: float64(whisperResponse.Duration)}, nil
}

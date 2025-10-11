package whispercppclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"m3u8dl_for_web/pkg/whisper"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

var _ whisper.WhisperHandler = &WhisperCppClient{}

type WhisperCppClient struct {
	InferenceUrl       string
	InferenceResultUrl string
	client             *http.Client
	splitDuration      int64
	debugMode          bool
	async              bool
}

func NewWhisperCppClient(config *WhisperCppClientConfig, httpClient *http.Client) *WhisperCppClient {
	cppClient := &WhisperCppClient{
		InferenceUrl:       config.InferenceUrl,
		InferenceResultUrl: config.InferenceResultUrl,
		splitDuration:      config.SplitDuration,
		debugMode:          config.DebugMode,
		client:             httpClient,
		async:              config.Async,
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

func (whisperCppClient *WhisperCppClient) handleWithSync(body []byte) (*WhisperResponse, error) {
	var whisperResponse = &WhisperResponse{}

	if err := json.Unmarshal(body, whisperResponse); err != nil {
		return nil, err
	}

	return whisperResponse, nil
}

func (whisperCppClient *WhisperCppClient) handleInferenceResult(taskId string) (*WhisperResponse, error) {
	var inferenceResultResponse = &InferenceResultResponse{}

	inferenceResultUrl, err := url.Parse(whisperCppClient.InferenceResultUrl)
	if err != nil {
		return nil, err
	}

	query := inferenceResultUrl.Query()
	query.Add("id", taskId)
	inferenceResultUrl.RawQuery = query.Encode()

	response, err := whisperCppClient.client.Get(inferenceResultUrl.String())
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error status:%d err:%s", response.StatusCode, body)
	}

	if err := json.Unmarshal(body, inferenceResultResponse); err != nil {
		return nil, err
	}
	logrus.Debugf("inferenceResultResponse %+v",inferenceResultResponse)

	if inferenceResultResponse.Status == "failed" {
		return nil, fmt.Errorf("inference result is failed:%s", inferenceResultResponse.Error)
	}

	if inferenceResultResponse.Status == "finished" {
		return inferenceResultResponse.Data, nil
	}

	if inferenceResultResponse.Status == "processing" {
		return nil, nil
	}

	return nil, fmt.Errorf("unknown status:%+v ", inferenceResultResponse)
}

func (whisperCppClient *WhisperCppClient) handleWithAsync(body []byte) (*WhisperResponse, error) {
	var (
		asyncInferenceResponse = &AsyncInferenceResponse{}
		interval               = 2 * time.Second
		resultChan             = make(chan InferenceResultResponse, 1)
		count                  = 0
		task                   func()
	)

	if err := json.Unmarshal(body, asyncInferenceResponse); err != nil {
		return nil, err
	}

	if asyncInferenceResponse.TaskId == "" {
		return nil, fmt.Errorf("inference task error: task id is empty")
	}

	logrus.Debugf("asyncInferenceResponse %+v",asyncInferenceResponse)

	task = func() {
		count++

		// 超过6小时，退出循环，报超时错误
		if count > 10800 {
			resultChan <- InferenceResultResponse{Error: fmt.Sprintf("timeout to waiting inference result ontask id '%s'", asyncInferenceResponse.TaskId)}
			return 
		}

		result, err := whisperCppClient.handleInferenceResult(asyncInferenceResponse.TaskId)
		if err != nil {
			resultChan <- InferenceResultResponse{Error: err.Error()}
			return
		}
		if result != nil {
			resultChan <- InferenceResultResponse{Data: result}
			return
		}

		time.AfterFunc(interval, task)
	}

	time.AfterFunc(0, task)

	result := <-resultChan
	if result.Error != "" {
		return nil, fmt.Errorf(result.Error)
	}

	return result.Data, nil
}

func (whisperCppClient *WhisperCppClient) HandleWhisper(ctx context.Context, input whisper.WhisperInput) (*whisper.WhisperOutput, error) {
	var (
		pr, pw = io.Pipe()
		writer = multipart.NewWriter(pw)
		req    = WhisperRequest{
			DebugMode:     whisperCppClient.debugMode,
			Prompt:        input.Prompt,
			Language:      input.Language,
			Temperature:   float64(input.Temperature),
			SplitOnWord:   true,
			MaxLen:        25,
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

	httpReq, err := http.NewRequestWithContext(ctx, "POST", whisperCppClient.InferenceUrl, pr)
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

	handleFunc := whisperCppClient.handleWithSync
	if whisperCppClient.async {
		handleFunc = whisperCppClient.handleWithAsync
	}

	if whisperResponse, err = handleFunc(body); err != nil {
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

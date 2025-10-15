package lateral

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/model/aggregate"
	"m3u8dl_for_web/pkg/whisper"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

var _ whisper.WhisperHandler = &LateralProvider{}

type LateralProvider struct {
	client *http.Client
	config *LateralProviderConfig
}

func NewLateralProvider(config *LateralProviderConfig, httpClient *http.Client) *LateralProvider {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	LateralProvider := &LateralProvider{
		client: httpClient,
		config: config,
	}

	return LateralProvider
}

// 8/16/16000 * splitDuration
func (lateralProvider *LateralProvider) MaximumFileSize() int64 {
	return 32000 * lateralProvider.config.SplitDuration
}

func (lateralProvider *LateralProvider) handleInferenceResult(ctx context.Context, taskId string) (*aggregate.SubtitleOutput, error) {
	var inferenceResultResponse = &LateralProviderResponse{}

	inferenceResultUrl, err := url.Parse(lateralProvider.config.InferenceResultUrl)
	if err != nil {
		return nil, err
	}

	query := inferenceResultUrl.Query()
	query.Add("id", taskId)
	inferenceResultUrl.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", inferenceResultUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	lateralProvider.config.AddBearerToken(httpReq)

	response, err := lateralProvider.client.Do(httpReq)
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
	logrus.Debugf("inferenceResultResponse %+v", inferenceResultResponse.Task)

	if inferenceResultResponse.Task == nil {
		return nil, fmt.Errorf("inference result is failed: %+v", inferenceResultResponse)
	}

	if inferenceResultResponse.Task.State == model.StateEnd {
		return &inferenceResultResponse.Task.Output, nil
	}
	if inferenceResultResponse.Task.State == model.StateReady {
		return nil, nil
	}

	return nil, fmt.Errorf("inference result is failed: %+v", inferenceResultResponse)

}

func (lateralProvider *LateralProvider) waitTaskFinish(ctx context.Context, taskId string) (*aggregate.SubtitleOutput, error) {

	type resultWrap struct {
		res *aggregate.SubtitleOutput
		err error
	}

	var (
		interval   = 2 * time.Second
		resultChan = make(chan resultWrap, 1)
		task       func()
	)

	task = func() {
		select {
		case <-ctx.Done():
			resultChan <- resultWrap{nil, ctx.Err()}
			return
		default:
		}
		result, err := lateralProvider.handleInferenceResult(ctx, taskId)
		if err != nil {
			resultChan <- resultWrap{err: err}
			return
		}
		if result != nil {
			resultChan <- resultWrap{res: result}
			return
		}

		time.AfterFunc(interval, task)
	}

	time.AfterFunc(0, task)

	result := <-resultChan
	if result.err != nil {
		return nil, result.err
	}

	return result.res, nil
}

func (lateralProvider *LateralProvider) HandleWhisper(ctx context.Context, input whisper.WhisperInput) (*whisper.WhisperOutput, error) {
	var (
		pr, pw = io.Pipe()
		writer = multipart.NewWriter(pw)
		req    = LateralProviderRequest{
			Provider:    lateralProvider.config.Provider,
			Prompt:      input.Prompt,
			Language:    input.Language,
			Temperature: float64(input.Temperature),
		}

		asyncInferenceResponse = &AsyncInferenceResponse{}
		subtitleOutput         *aggregate.SubtitleOutput
	)

	logrus.Debugf(" lateral req:%+v", req)

	go func() {
		_ = req.WriteMultipartWithFile(writer, "file", input.FilePath)
		writer.Close()
		pw.Close()
	}()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", lateralProvider.config.InferenceUrl, pr)
	if err != nil {
		return nil, err
	}
	lateralProvider.config.AddBearerToken(httpReq)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := lateralProvider.client.Do(httpReq)
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

	if err := json.Unmarshal(body, asyncInferenceResponse); err != nil {
		return nil, err
	}

	if asyncInferenceResponse.Err != "" {
		return nil, fmt.Errorf("inference task error: %s", asyncInferenceResponse.Err)
	}

	if asyncInferenceResponse.TaskId == "" {
		return nil, fmt.Errorf("inference task error: task id is empty")
	}

	logrus.Debugf("asyncInferenceResponse %+v", asyncInferenceResponse)

	timeoutCtx, cancel := context.WithTimeout(ctx, lateralProvider.config.GetTaskTimeoutDurationOrDefault())
	defer cancel()
	if subtitleOutput, err = lateralProvider.waitTaskFinish(timeoutCtx, asyncInferenceResponse.TaskId); err != nil {
		return nil, err
	}

	var segments = make([]whisper.Segment, len(subtitleOutput.SegmentList))
	for i, segment := range subtitleOutput.SegmentList {
		segments[i] = whisper.Segment{
			Num:   int(segment.Num),
			Start: segment.Start,
			End:   segment.End,
			Text:  segment.Text,
		}

	}

	return &whisper.WhisperOutput{Segments: segments, Duration: float64(subtitleOutput.MediaDuration)}, nil
}

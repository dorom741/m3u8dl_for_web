package lateral

import (
	"fmt"
	"io"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/model/aggregate"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type LateralProviderConfig struct {
	InferenceUrl        string `yaml:"inferenceUrl"`
	InferenceResultUrl  string `yaml:"inferenceResultUrl"`
	SplitDuration       int64  `yaml:"splitDuration"`
	TaskTimeoutDuration string `yaml:"taskTimeoutDuration"`
	Provider            string `yaml:"provider"`
	BearerToken         string `yaml:"bearerToken"`

	// DebugMode          bool   `yaml:"debugMode"`
	// Async              bool   `yaml:"async"`
}

func (config *LateralProviderConfig) GetTaskTimeoutDurationOrDefault() time.Duration {
	duration, err := time.ParseDuration(config.TaskTimeoutDuration)
	if err != nil || duration.Seconds() == 0 {
		duration = time.Hour * 12
	}

	return duration
}

func (config *LateralProviderConfig) AddBearerToken(req *http.Request) {
	if config.BearerToken == "" {
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.BearerToken))
}

type LateralProviderRequest struct {
	// FilePath string `json:"filepath"`
	Provider    string  `json:"provider"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	Language    string  `json:"language"`
}

// ToValues 将结构体转换为 url.Values（用于 application/x-www-form-urlencoded 或 multipart 的非文件字段）
func (r *LateralProviderRequest) ToValues() url.Values {
	v := url.Values{}
	v.Set("provider", r.Provider)
	v.Set("offset_n", r.Prompt)
	v.Set("temperature", strconv.FormatFloat(r.Temperature, 'f', -1, 64))
	v.Set("language", r.Language)

	return v
}

// WriteMultipartWithFile 写 multipart 表单，包含 file 字段（fileFieldName），并把其它字段写入 form 字段。
// filePath 可传空以只写字段不写文件。
func (r *LateralProviderRequest) WriteMultipartWithFile(w *multipart.Writer, fileFieldName, filePath string) error {
	// 先写文件字段（如果提供）
	if filePath != "" {
		fw, err := w.CreateFormFile(fileFieldName, filepath.Base(filePath))
		if err != nil {
			return err
		}
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(fw, f); err != nil {
			return err
		}
	}

	// 写其它字段
	vals := r.ToValues()
	for key, valsArr := range vals {
		for _, vv := range valsArr {
			if err := w.WriteField(key, vv); err != nil {
				return err
			}
		}
	}
	return nil
}

type AsyncInferenceResponse struct {
	Err    string `json:"err"`
	TaskId string `json:"task_id"`
}

type LateralProviderResponse struct {
	Err  string `json:"err"`
	Task *model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]
}

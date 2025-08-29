package whispercppclient

import (
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

type WhisperCppClientConfig struct {
	BaseUrl string `yaml:"baseUrl"`
	SplitDuration int64 `yaml:"splitDuration"`
	DebugMode bool `yaml:"debugMode"`

}

// WhisperRequest 对应 server 中的可接受参数（简化与合理推断类型）
type WhisperRequest struct {
	OffsetT    int `form:"offset_t"`
	OffsetN    int `form:"offset_n"`
	Duration   int `form:"duration"`
	MaxContext int `form:"max_context"`
	MaxLen     int `form:"max_len"`
	BestOf     int `form:"best_of"`
	BeamSize   int `form:"beam_size"`
	AudioCtx   int `form:"audio_ctx"`

	WordThold      float64 `form:"word_thold"`
	EntropyThold   float64 `form:"entropy_thold"`
	LogprobThold   float64 `form:"logprob_thold"`
	Temperature    float64 `form:"temperature"`
	TemperatureInc float64 `form:"temperature_inc"`
	SuppressNST    bool    `form:"suppress_nst"` // server 同时支持 suppress_non_speech/suppress_nst

	DebugMode               bool `form:"debug_mode"`
	Translate               bool `form:"translate"`
	DetectLanguage          bool `form:"detect_language"`
	Diarize                 bool `form:"diarize"`
	Tinydiarize             bool `form:"tinydiarize"`
	SplitOnWord             bool `form:"split_on_word"`
	NoFallback              bool `form:"no_fallback"`
	PrintSpecial            bool `form:"print_special"`
	PrintColors             bool `form:"print_colors"`
	PrintRealtime           bool `form:"print_realtime"`
	PrintProgress           bool `form:"print_progress"`
	NoTimestamps            bool `form:"no_timestamps"`
	NoContext               bool `form:"no_context"`
	NoLanguageProbabilities bool `form:"no_language_probabilities"`

	Language       string `form:"language"`
	Prompt         string `form:"prompt"`
	ResponseFormat string `form:"response_format"`

	Vad                     bool    `form:"vad"`
	VadModel                string  `form:"vad_model"`
	VadThreshold            float64 `form:"vad_threshold"`
	VadMinSpeechDurationMs  int     `form:"vad_min_speech_duration_ms"`
	VadMinSilenceDurationMs int     `form:"vad_min_silence_duration_ms"`
	VadMaxSpeechDurationS   float64 `form:"vad_max_speech_duration_s"`
	VadSpeechPadMs          int     `form:"vad_speech_pad_ms"`
	VadSamplesOverlap       float64 `form:"vad_samples_overlap"`
}

// ToValues 将结构体转换为 url.Values（用于 application/x-www-form-urlencoded 或 multipart 的非文件字段）
func (r *WhisperRequest) ToValues() url.Values {
	v := url.Values{}

	// 整数型
	v.Set("offset_t", strconv.Itoa(r.OffsetT))
	v.Set("offset_n", strconv.Itoa(r.OffsetN))
	v.Set("duration", strconv.Itoa(r.Duration))
	v.Set("max_context", strconv.Itoa(r.MaxContext))
	v.Set("max_len", strconv.Itoa(r.MaxLen))
	v.Set("best_of", strconv.Itoa(r.BestOf))
	v.Set("beam_size", strconv.Itoa(r.BeamSize))
	v.Set("audio_ctx", strconv.Itoa(r.AudioCtx))

	// 浮点型
	v.Set("word_thold", strconv.FormatFloat(r.WordThold, 'f', -1, 64))
	v.Set("entropy_thold", strconv.FormatFloat(r.EntropyThold, 'f', -1, 64))
	v.Set("logprob_thold", strconv.FormatFloat(r.LogprobThold, 'f', -1, 64))
	v.Set("temperature", strconv.FormatFloat(r.Temperature, 'f', -1, 64))
	v.Set("temperature_inc", strconv.FormatFloat(r.TemperatureInc, 'f', -1, 64))
	v.Set("vad_threshold", strconv.FormatFloat(r.VadThreshold, 'f', -1, 64))
	v.Set("vad_max_speech_duration_s", strconv.FormatFloat(r.VadMaxSpeechDurationS, 'f', -1, 64))
	v.Set("vad_samples_overlap", strconv.FormatFloat(r.VadSamplesOverlap, 'f', -1, 64))

	// 布尔型（以 "true"/"false" 字符串传）
	v.Set("suppress_nst", strconv.FormatBool(r.SuppressNST))
	v.Set("debug_mode", strconv.FormatBool(r.DebugMode))
	v.Set("translate", strconv.FormatBool(r.Translate))
	v.Set("detect_language", strconv.FormatBool(r.DetectLanguage))
	v.Set("diarize", strconv.FormatBool(r.Diarize))
	v.Set("tinydiarize", strconv.FormatBool(r.Tinydiarize))
	v.Set("split_on_word", strconv.FormatBool(r.SplitOnWord))
	v.Set("no_fallback", strconv.FormatBool(r.NoFallback))
	v.Set("print_special", strconv.FormatBool(r.PrintSpecial))
	v.Set("print_colors", strconv.FormatBool(r.PrintColors))
	v.Set("print_realtime", strconv.FormatBool(r.PrintRealtime))
	v.Set("print_progress", strconv.FormatBool(r.PrintProgress))
	v.Set("no_timestamps", strconv.FormatBool(r.NoTimestamps))
	v.Set("no_context", strconv.FormatBool(r.NoContext))
	v.Set("no_language_probabilities", strconv.FormatBool(r.NoLanguageProbabilities))
	v.Set("vad", strconv.FormatBool(r.Vad))

	// 字符串
	v.Set("language", r.Language)
	v.Set("prompt", r.Prompt)
	v.Set("response_format", r.ResponseFormat)
	v.Set("vad_model", r.VadModel)

	// VAD 整数
	v.Set("vad_min_speech_duration_ms", strconv.Itoa(r.VadMinSpeechDurationMs))
	v.Set("vad_min_silence_duration_ms", strconv.Itoa(r.VadMinSilenceDurationMs))
	v.Set("vad_speech_pad_ms", strconv.Itoa(r.VadSpeechPadMs))

	return v
}

// WriteMultipartWithFile 写 multipart 表单，包含 file 字段（fileFieldName），并把其它字段写入 form 字段。
// filePath 可传空以只写字段不写文件。
func (r *WhisperRequest) WriteMultipartWithFile(w *multipart.Writer, fileFieldName, filePath string) error {
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

type WhisperResponse struct {
	Task                        string                `json:"task"`
	Language                    string                `json:"language"`
	Duration                    float64               `json:"duration"`
	Text                        string                `json:"text"`
	Segments                    []Segment             `json:"segments"`
	DetectedLanguage            string                `json:"detected_language"`
	DetectedLanguageProbability float64               `json:"detected_language_probability"`
	LanguageProbabilities       LanguageProbabilities `json:"language_probabilities"`
}

type LanguageProbabilities struct {
	En  float64 `json:"en"`
	Ar  float64 `json:"ar"`
	Ur  float64 `json:"ur"`
	Mi  float64 `json:"mi"`
	Yi  float64 `json:"yi"`
	Haw float64 `json:"haw"`
}

type Segment struct {
	ID           int64   `json:"id"`
	Text         string  `json:"text"`
	Start        float64 `json:"start"`
	End          float64 `json:"end"`
	Tokens       []int64 `json:"tokens"`
	Words        []Word  `json:"words"`
	Temperature  float64 `json:"temperature"`
	AvgLogprob   float64 `json:"avg_logprob"`
	NoSpeechProb float64 `json:"no_speech_prob"`
}

type Word struct {
	Word        string  `json:"word"`
	Start       float64 `json:"start"`
	End         float64 `json:"end"`
	TDtw        int64   `json:"t_dtw"`
	Probability float64 `json:"probability"`
}

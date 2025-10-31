package translation

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"golang.org/x/time/rate"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sirupsen/logrus"
)

var _ ITranslation = &OpenAiCompatibleTranslation{}

type OpenAiCompatibleTranslation struct {
	config   *OpenAiCompatibleConfig
	client   openai.Client
	messages []openai.ChatCompletionMessageParamUnion
}

func NewOpenAiCompatibleTranslation(config *OpenAiCompatibleConfig, httpClient *http.Client) *OpenAiCompatibleTranslation {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	opts := []option.RequestOption{
		option.WithAPIKey(config.ApiKey),
		option.WithBaseURL(config.BaseUrl),
		option.WithHTTPClient(httpClient),
	}

	if config.RPM > 0 {
		burst := (config.RPM + 60 - 1) / 60
		limiter := rate.NewLimiter(rate.Every(time.Second*time.Duration(int(60/config.RPM)+1)), burst)
		rateLimitMiddleware := func(request *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			if err := limiter.Wait(context.Background()); err != nil {
				return nil, err
			}
			return next(request)
		}
		logrus.Infof("enable openAi compatible translation rate limit RPM %d burst %d", config.RPM, burst)

		opts = append(opts, option.WithMiddleware(rateLimitMiddleware))
	}
	client := openai.NewClient(opts...)

	return &OpenAiCompatibleTranslation{
		config: config,
		client: client,
		messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(config.SystemPrompt),
		},
	}
}

func (translation *OpenAiCompatibleTranslation) GetName() string {
	return "OpenAiCompatibleTranslation"
}

func (translation *OpenAiCompatibleTranslation) SupportMultipleTextBySeparator() (bool, string) {
	return len(translation.config.MultipleTextSeparator) > 0, translation.config.MultipleTextSeparator
}

func (translation *OpenAiCompatibleTranslation) Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	promptTmpl, err := template.New("translate").Parse(translation.config.Prompt)
	if err != nil {
		return "", err
	}

	directiveBuffer := new(bytes.Buffer)
	data := map[string]string{
		"text":       text,
		"sourceLang": sourceLang,
		"targetLang": targetLang,
	}

	if err := promptTmpl.Execute(directiveBuffer, data); err != nil {
		return "", err
	}

	// logrus.Debugf("translate prompt %s", directiveBuffer.String())

	translation.messages = append(translation.messages, openai.UserMessage(directiveBuffer.String()))

	chatCompletion, err := translation.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(translation.config.Model),
		Messages: translation.messages,
	})
	if err != nil {
		return "", err
	}
	logrus.Debugf("openAi compatible translation chatCompletion Choices %+v", chatCompletion)
	if len(chatCompletion.Choices) == 0 {
		return "", fmt.Errorf("openAi compatible translation: no choices returned")
	}
	result := chatCompletion.Choices[0].Message.Content
	translation.messages = append(translation.messages, chatCompletion.Choices[0].Message.ToParam())

	messagesLen := len(translation.messages)
	if messagesLen > translation.config.ContextLen+1 {
		start := messagesLen - translation.config.ContextLen - 1
		translation.messages = append([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(translation.config.SystemPrompt),
		}, translation.messages[start:]...)
	}

	words := "</think>"
	if idx := strings.Index(result, words); idx != -1 {
		// 获取 </think> 后面的内容
		result = result[idx+len(words):]
	}

	result = strings.TrimSpace(result)                                  // 去掉开头和结尾的空白字符
	result = strings.TrimSuffix(strings.TrimPrefix(result, "\n"), "\n") // 去掉开头和结尾的换行字符
	return result, nil
}

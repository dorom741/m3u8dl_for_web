package translation

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	"m3u8dl_for_web/conf"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sirupsen/logrus"
)

var _ ITranslation = &OpenAiCompatibleTranslation{}

type OpenAiCompatibleTranslation struct {
	config   *conf.OpenAiCompatibleConfig
	client   *openai.Client
	messages []openai.ChatCompletionMessageParamUnion
}

func NewOpenAiCompatibleTranslation(config *conf.OpenAiCompatibleConfig) *OpenAiCompatibleTranslation {
	client := openai.NewClient(
		option.WithAPIKey(config.ApiKey),
		option.WithBaseURL(config.BaseUrl),
	)

	return &OpenAiCompatibleTranslation{
		config: config,
		client: client,
		messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(config.SystemPrompt),
		},
	}
}

func (translation *OpenAiCompatibleTranslation) SupportMultipleTextBySeparator() (bool, string) {
	return false, "\n"
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

	logrus.Infof("translate prompt %s", directiveBuffer.String())

	translation.messages = append(translation.messages, openai.UserMessage(directiveBuffer.String()))

	chatCompletion, err := translation.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.F(translation.config.Model),
		Messages: openai.F(translation.messages),
	})
	if err != nil {
		panic(err)
	}
	logrus.Infof("chatCompletion Choices %+v", chatCompletion)
	result := chatCompletion.Choices[0].Message.Content
	translation.messages = append(translation.messages, chatCompletion.Choices[0].Message)

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
		result = strings.TrimSpace(result) // 去掉开头和结尾的空白字符
	}

	return result, nil
}

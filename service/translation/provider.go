package translation

import "context"

type ITranslation interface {
	Translate(ctx context.Context, text string, sourceLang string, targetLang string) (string, error)
	SupportMultipleTextByPunctuation() (bool, string)
}

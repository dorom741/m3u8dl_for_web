package translation

import (
	"context"
	"testing"

	"m3u8dl_for_web/conf"
)

func init() {
	conf.InitConf("../../config.yaml")
}

func TestDeelxTranslate(t *testing.T) {
	ctx := context.Background()

	translation := NewDeepLXTranslation(conf.ConfigInstance.Translation.DeeplX.Url, nil)

	result, err := translation.Translate(ctx, "car", "en", "zh")
	if err != nil {
		t.Error(err)
	}
	t.Logf("translate resultL %+v", result)
}

func TestOpenaiTranslate(t *testing.T) {
	ctx := context.Background()

	translation := NewOpenAiCompatibleTranslation(conf.ConfigInstance.Translation.OpenAiCompatible)

	for i, word := range []string{"computer", "car", "bye", "page", "big", "sad", "hello", "goodbye", "happy", "love", "hate", "like", "dislike", "good", "bad", "beautiful"} {
		result, err := translation.Translate(ctx, word, "en", "zh")
		if err != nil {
			t.Error(err)
		}
		t.Logf("translate %d resultL %+v", i, result)

	}

	// result, err := translation.Translate(ctx, "computer", "en", "zh")
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Logf("translate resultL %+v", result)
}

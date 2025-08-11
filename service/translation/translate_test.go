package translation

import (
	"context"
	"testing"
	"time"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
)

func init() {
	conf.InitConf("../../config.yaml")
	if err := infra.InitHttpClientWithProxy(conf.ConfigInstance.Server.HttpClientProxy); err != nil {
		panic(err)
	}
}

func TestDeepLXTranslate(t *testing.T) {
	ctx := context.Background()

	translation := NewDeepLXTranslation(conf.ConfigInstance.Translation.DeepLX, infra.DefaultHttpClient)

	for i, word := range []string{"computer", "car", "bye", "page", "big", "sad", "hello", "goodbye", "happy", "love", "hate", "like", "dislike", "good", "bad", "beautiful"} {
		result, err := translation.Translate(ctx, word, "en", "zh")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("[%s] translate %d result %+v",time.Now().Format(time.RFC3339), i, result)

	}
}

func TestOpenaiTranslate(t *testing.T) {
	ctx := context.Background()

	translation := NewOpenAiCompatibleTranslation(conf.ConfigInstance.Translation.OpenAiCompatible)

	for i, word := range []string{"computer", "car", "bye", "page", "big", "sad", "hello", "goodbye", "happy", "love", "hate", "like", "dislike", "good", "bad", "beautiful"} {
		result, err := translation.Translate(ctx, word, "en", "zh")
		if err != nil {
			t.Error(err)
		}
		t.Logf("[%s] translate %d result %+v",time.Now().Format(time.RFC3339), i, result)

	}

	// result, err := translation.Translate(ctx, "computer", "en", "zh")
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Logf("translate resultL %+v", result)
}

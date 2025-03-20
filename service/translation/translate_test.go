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

	result, err := translation.Translate(ctx, "car", "en", "zh")
	if err != nil {
		t.Error(err)
	}
	t.Logf("translate resultL %+v", result)
}

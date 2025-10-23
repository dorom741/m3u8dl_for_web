package test

import (
	"context"
	"strings"
	"testing"

	"m3u8dl_for_web/service"

	"github.com/sirupsen/logrus"
)

var translateFragmentList = []string{
	"He bore the cross and went to the bank",                           //他背着十字架走向河岸
	"The pitcher will pitch on the pitch after the pitch.",             //投手将在场地准备工作后开始投球
	"She left the rest on the left in the jar",                         //她把左边剩下的东西留在了罐子里
	"The wind will wind around the windmill as we wind up the project", //当我们结束这个项目时，风会绕着风车吹
	"goodbye", "happy", "love", "hate", "like", "dislike", "good", "bad", "beautiful",
}

func TestTranslateHub(t *testing.T) {
	ctx := context.Background()
	translation := service.TranslationServiceInstance

	for i, word := range translateFragmentList {
		result, err := translation.Translate(ctx, word, "en", "zh")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("translate %d result %+v", i, result)

	}
}

func TestTranslateHubBatchTranslate(t *testing.T) {
	ctx := context.Background()
	translation := service.TranslationServiceInstance

	result, err := translation.BatchTranslate(ctx, translateFragmentList, "en", "zh")
	if err != nil {
		logrus.Error(err.Error())
		t.Error(err)
	}

	logrus.Infof("batch translate result '%+v'", strings.Join(result, "','"))
}

// func TestTranslate(t *testing.T) {
// 	ctx := context.Background()

// 	translation := translation.NewDeepLXTranslation(&translation.DeepLXConfig{}, infra.DefaultHttpClient)

// 	for i, word := range translateFragmentList {
// 		result, err := translation.Translate(ctx, word, "en", "zh")
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		t.Logf("[%s] translate %d result %+v", time.Now().Format(time.RFC3339), i, result)

// 	}
// }

// func TestOpenaiTranslate(t *testing.T) {
// 	ctx := context.Background()

// 	translation := translation.NewOpenAiCompatibleTranslation(&translation.OpenAiCompatibleConfig{})

// 	for i, word := range translateFragmentList {
// 		result, err := translation.Translate(ctx, word, "en", "zh")
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		t.Logf("[%s] translate %d result %+v", time.Now().Format(time.RFC3339), i, result)

// 	}

// 	// result, err := translation.Translate(ctx, "computer", "en", "zh")
// 	// if err != nil {
// 	// 	t.Error(err)
// 	// }
// 	// t.Logf("translate resultL %+v", result)
// }

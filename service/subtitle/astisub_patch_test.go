package subtitle

import (
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	sentence := []string{
		"Parlez-vous français? ",
		"Ich spreche Französisch nur ein bisschen. ",
		"A little bit is better than nothing.",
		"你好",
	}

	for _, item := range sentence {
		detectLanguage(item)
	}
}

func TestSplitBilingualSubtitle(t *testing.T) {
	result, err := SplitBilingualSubtitle("resource/samples/jfk.ass")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result %+v", result)
}

package whisper

import (
	"testing"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

func TestProcess(t *testing.T) {
	modelPath := "/workplace/project/demo/m3u8dl_for_web/resource/download/ggml-base.bin"
	inputPath := "/workplace/project/demo/whisper.cpp/samples/jfk.mp3"

	params := Params{
		IsTokenize: true,
	}

	if err := params.InputFromFile(inputPath); err != nil {
		t.Fatal(err)
	}

	model, err := whisper.New(modelPath)
	if err != nil {
		t.Fatal(err)
	}
	defer model.Close()

	// Create processing context
	context, err := model.NewContext()
	if err != nil {
		t.Fatal(err)
	}

	segmentList, err := Process(context, params)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("segmentList:%+v", segmentList)
}

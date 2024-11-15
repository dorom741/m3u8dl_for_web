package sherpa

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/pkg/whisper"
)

var _ whisper.WhisperHandler = &SherpaWhisper{}

type SherpaWhisper struct {
	// vadModelPath       string

	embeddingModelPath string
	pyannoteModelPath  string

	modelConfig sherpa.OfflineModelConfig
}

func NewSherpaWhisper(sherpaConfig SherpaConfig) *SherpaWhisper {
	return &SherpaWhisper{
		// vadModelPath:       vadModelPath,
		embeddingModelPath: sherpaConfig.EmbeddingModelPath,
		pyannoteModelPath:  sherpaConfig.PyannoteModelPath,
		modelConfig:        sherpaConfig.OfflineModelConfig,
	}
}

// 最大 25MB
func (sherpaWhisper *SherpaWhisper) MaximumFileSize() int64 {
	return 26214400
}

func (sherpaWhisper *SherpaWhisper) HandleWhisper(ctx context.Context, input whisper.WhisperInput) (*whisper.WhisperOutput, error) {
	var progressCallback  = func (int)  {}
	if input.ProgressCallback != nil {
		progressCallback = input.ProgressCallback
	}

	file, err := os.Open(input.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dec, pcmBuffer, err := sherpaWhisper.readPCMInfo(file)
	if err != nil {
		return nil, err
	}

	duration, err := dec.Duration()
	if err != nil {
		return nil, err
	}

	speakerDiarizationSegmentList, err := sherpaWhisper.SpeakerDiarization(pcmBuffer.AsFloat32Buffer().Data)
	if err != nil {
		return nil, err
	}
	whisperSegmentLen :=  len(speakerDiarizationSegmentList)
	whisperSegments := make([]whisper.Segment,whisperSegmentLen)
	logrus.Debugf("speaker diarization segment list:%+v", speakerDiarizationSegmentList)
	progressCallback(10)

	recognizerConfig := sherpaWhisper.newRecognizerConfig()
	recognizer := sherpa.NewOfflineRecognizer(recognizerConfig)
	defer sherpa.DeleteOfflineRecognizer(recognizer)

	for i, segment := range speakerDiarizationSegmentList {
		pcmData, err := sherpaWhisper.selectPCMData(dec.SampleRate, pcmBuffer, float64(segment.Start), float64(segment.End))
		if err != nil {
			return nil, err
		}

		result := sherpaWhisper.OfflineRecognizer(recognizer, int(dec.SampleRate), pcmData)
		if result == nil {
			continue
		}

		progressCallback((i+1)*90/whisperSegmentLen+10)

		whisperSegments[i] = whisper.Segment{
			Num:   i,
			Start: float64(segment.Start),
			End:   float64(segment.End),
			Text:  result.Text,
		}

	}

	return &whisper.WhisperOutput{Duration: duration.Seconds(), Segments: whisperSegments}, nil
}

func (sherpaWhisper *SherpaWhisper) SpeakerDiarization(inputdata []float32) ([]sherpa.OfflineSpeakerDiarizationSegment, error) {
	c := sherpa.OfflineSpeakerDiarizationConfig{}
	c.Segmentation.Pyannote.Model = sherpaWhisper.pyannoteModelPath
	c.Segmentation.NumThreads = sherpaWhisper.modelConfig.NumThreads
	c.Segmentation.Debug = sherpaWhisper.modelConfig.Debug

	c.Embedding.Model = sherpaWhisper.embeddingModelPath
	c.Embedding.NumThreads = sherpaWhisper.modelConfig.NumThreads
	c.Embedding.Debug = sherpaWhisper.modelConfig.Debug

	// The test wave file contains 4 speakers, so we use 4 here
	// c.Clustering.NumClusters = 4

	// if you don't know the actual numbers in the wave file,
	// then please don't set NumClusters; you need to use
	//
	// config.Clustering.Threshold = 0.5
	//

	// A larger Threshold leads to fewer clusters
	// A smaller Threshold leads to more clusters

	sd := sherpa.NewOfflineSpeakerDiarization(&c)
	defer sherpa.DeleteOfflineSpeakerDiarization(sd)

	segments := sd.Process(inputdata)
	return segments, nil
}

func (sherpaWhisper *SherpaWhisper) OfflineRecognizer(recognizer *sherpa.OfflineRecognizer, sampleRate int, inputdata []float32) *sherpa.OfflineRecognizerResult {
	stream := sherpa.NewOfflineStream(recognizer)
	defer sherpa.DeleteOfflineStream(stream)
	stream.AcceptWaveform(sampleRate, inputdata)
	recognizer.Decode(stream)
	result := stream.GetResult()

	logrus.Debugf("offline recognizer result:  %+v", result)

	return result
}

func (sherpaWhisper *SherpaWhisper) newRecognizerConfig() *sherpa.OfflineRecognizerConfig {
	recognizerConfig := &sherpa.OfflineRecognizerConfig{}
	recognizerConfig.FeatConfig.SampleRate = 16000
	recognizerConfig.FeatConfig.FeatureDim = 80

	recognizerConfig.ModelConfig = sherpaWhisper.modelConfig

	// recognizerConfig.ModelConfig.NumThreads = 4
	// recognizerConfig.ModelConfig.Debug = 1
	// recognizerConfig.ModelConfig.Provider = "cpu"

	return recognizerConfig
}

func (sherpaWhisper *SherpaWhisper) readPCMInfo(reader io.ReadSeeker) (*wav.Decoder, *audio.IntBuffer, error) {
	dec := wav.NewDecoder(reader)

	buffer, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, nil, err
	}
	if dec.SampleRate != 16000 {
		return nil, nil, fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	}
	if dec.NumChans != 1 {
		return nil, nil, fmt.Errorf("unsupported number of channels: %d", dec.NumChans)
	}

	if _, err := dec.Seek(0, 0); err != nil {
		return nil, nil, err
	}
	// duration, err := dec.Duration()
	// if err != nil {
	// 	return nil,  err
	// }

	return dec, buffer, nil
}

func (sherpaWhisper *SherpaWhisper) selectPCMData(sampleRate uint32, audioBuffer *audio.IntBuffer, startTime float64, endTime float64) ([]float32, error) {
	startSample := int(startTime * float64(sampleRate))
	endSample := int(endTime * float64(sampleRate))

	if startSample < 0 || endSample > len(audioBuffer.Data) || startSample >= endSample {
		return nil, fmt.Errorf("Invalid start or end time")
	}

	trimmedBuffer := &audio.IntBuffer{
		Format:         audioBuffer.Format,
		SourceBitDepth: audioBuffer.SourceBitDepth,
		Data:           audioBuffer.Data[startSample:endSample],
	}

	return trimmedBuffer.AsFloat32Buffer().Data, nil
}

package sherpa

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/pkg/whisper"
)

var _ whisper.WhisperHandler = &SherpaWhisper{}

type SherpaWhisper struct {
	vadModelPath            string
	whisperDecoderModelPath string
	whisperEncoderModelPath string
	whisperModelTokensPath  string
	embeddingModelPath      string
	pyannoteModelPath       string
}

func NewSherpaWhisper(vadModelPath string, whisperDecoderModelPath string, whisperEncoderModelPath string, whisperModelTokensPath string, embeddingModelPath string, pyannoteModelPath string) *SherpaWhisper {
	return &SherpaWhisper{
		vadModelPath:            vadModelPath,
		whisperDecoderModelPath: whisperDecoderModelPath,
		whisperEncoderModelPath: whisperEncoderModelPath,
		whisperModelTokensPath:  whisperModelTokensPath,
		embeddingModelPath:      embeddingModelPath,
		pyannoteModelPath:       pyannoteModelPath,
	}
}

func (sherpaWhisper *SherpaWhisper) MaximumFileSize() int64 {
	return 0
}

func (sherpaWhisper *SherpaWhisper) HandleWhisper(ctx context.Context, input whisper.WhisperInput) (*whisper.WhisperOutput, error) {
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

	segmentList, err := sherpaWhisper.SpeakerDiarization(pcmBuffer.AsFloat32Buffer().Data)
	if err != nil {
		return nil, err
	}

	whisperSegments := make([]whisper.Segment, len(segmentList))
	logrus.Println(segmentList)

	recognizerConfig := sherpaWhisper.newRecognizerConfig()
	recognizer := sherpa.NewOfflineRecognizer(recognizerConfig)
	defer sherpa.DeleteOfflineRecognizer(recognizer)
	stream := sherpa.NewOfflineStream(recognizer)

	for i, segment := range segmentList {
		pcmData, err := sherpaWhisper.selectPCMData(dec.SampleRate, pcmBuffer, float64(segment.Start), float64(segment.End))
		if err != nil {
			return nil, err
		}
		stream.AcceptWaveform(recognizerConfig.FeatConfig.SampleRate, pcmData)
		recognizer.Decode(stream)
		result := stream.GetResult()

		logrus.Infof("result:  %+v", result)

		whisperSegments[i] = whisper.Segment{
			Num:   i,
			Start: float64(segment.Start),
			End:   float64(segment.End),
			Text:  result.Text,
		}
	}

	// logrus.Println("Emotion: " + result.Emotion)
	// logrus.Println("Lang: " + result.Lang)
	// logrus.Println("Event: " + result.Event)
	// logrus.Printf("Timestamp: %v", result.Timestamps)
	// logrus.Printf("Tokens: %v", result.Tokens)
	// logrus.Printf("Wave duration: %v seconds", duration)

	// vad.AcceptWaveform(data)
	// for !vad.IsEmpty() {
	// 	speechSegment := vad.Front()
	// 	vad.Pop()

	// 	fmt.Printf("speechSegment.Start: %v\n", speechSegment.Start)

	// 	audio := &sherpa.Wave{}
	// 	audio.Samples = speechSegment.Samples
	// 	audio.SampleRate = config.SampleRate

	// 	// Now decode it
	// 	stream := sherpa.NewOfflineStream(recognizer)
	// 	defer sherpa.DeleteOfflineStream(stream)
	// 	stream.AcceptWaveform(audio.SampleRate, audio.Samples)
	// 	recognizer.Decode(stream)
	// 	result := stream.GetResult()
	// 	text := strings.ToLower(result.Text)
	// 	text = strings.Trim(text, " ")

	// }

	return &whisper.WhisperOutput{Segments: whisperSegments, Duration: duration.Seconds()}, nil
}

// Please download silero_vad.onnx from
// https://github.com/snakers4/silero-vad/raw/master/src/silero_vad/data/silero_vad.onnx
func (sherpaWhisper *SherpaWhisper) WithVad() {
	config := sherpa.VadModelConfig{}

	config.SileroVad.Model = sherpaWhisper.vadModelPath
	config.SileroVad.Threshold = 0.5
	config.SileroVad.MinSilenceDuration = 0.5
	config.SileroVad.MinSpeechDuration = 0.25
	config.SileroVad.WindowSize = 512
	config.SileroVad.MaxSpeechDuration = 5
	// config.SampleRate = 16000
	// config.NumThreads = 1
	// config.Provider = "cpu"
	// config.Debug = 1
	var bufferSizeInSeconds float32 = 20
	vad := sherpa.NewVoiceActivityDetector(&config, bufferSizeInSeconds)
	defer sherpa.DeleteVoiceActivityDetector(vad)
}

func (sherpaWhisper *SherpaWhisper) SpeakerDiarization(inputdata []float32) ([]sherpa.OfflineSpeakerDiarizationSegment, error) {
	c := sherpa.OfflineSpeakerDiarizationConfig{}
	c.Segmentation.Pyannote.Model = sherpaWhisper.pyannoteModelPath
	c.Embedding.Model = sherpaWhisper.embeddingModelPath

	c.Clustering.NumClusters = runtime.NumCPU()
	// if you don't know the actual numbers in the wave file,
	// then please don't set NumClusters; you need to use
	//
	c.Clustering.Threshold = 0.5

	// A larger Threshold leads to fewer clusters
	// A smaller Threshold leads to more clusters

	sd := sherpa.NewOfflineSpeakerDiarization(&c)
	defer sherpa.DeleteOfflineSpeakerDiarization(sd)

	segments := sd.Process(inputdata)
	return segments, nil
}

func (sherpaWhisper *SherpaWhisper) newRecognizerConfig() *sherpa.OfflineRecognizerConfig {
	recognizerConfig := &sherpa.OfflineRecognizerConfig{}
	recognizerConfig.FeatConfig.SampleRate = 16000
	recognizerConfig.FeatConfig.FeatureDim = 80
	recognizerConfig.ModelConfig.Whisper.Encoder = sherpaWhisper.whisperEncoderModelPath
	recognizerConfig.ModelConfig.Whisper.Decoder = sherpaWhisper.whisperDecoderModelPath
	recognizerConfig.ModelConfig.Tokens = sherpaWhisper.whisperModelTokensPath
	recognizerConfig.ModelConfig.NumThreads = 4
	recognizerConfig.ModelConfig.Debug = 1
	recognizerConfig.ModelConfig.Provider = "cuda"

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

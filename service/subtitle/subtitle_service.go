package subtitle

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/model"

	"m3u8dl_for_web/pkg/media"
	"m3u8dl_for_web/pkg/whisper"
	"m3u8dl_for_web/service/translation"
)

type SubtitleService struct {
	tempAudioPath string
	translation   translation.ITranslation
	cache         *infra.FileCache
}

func NewSubtitleService(tempAudioPath string, cache *infra.FileCache, translation translation.ITranslation) *SubtitleService {
	RegisterWhisperProvider()

	return &SubtitleService{
		tempAudioPath: tempAudioPath,
		translation:   translation,
		cache:         cache,
	}
}

func (service *SubtitleService) cacheKey(provider string, input whisper.WhisperInput) string {
	prefix := "subtitle_service"

	filename := filepath.Base(input.FilePath)
	ext := filepath.Ext(input.FilePath)

	return fmt.Sprintf("%s_%s_%s_%.2f", prefix, filename[:len(filename)-len(ext)], provider, input.Temperature)
}

func (service *SubtitleService) GenerateSubtitle(ctx context.Context, input model.SubtitleInput) error {
	var (
		filename     = filepath.Base(input.InputPath)
		ext          = filepath.Ext(filename)
		pureFilename = strings.ReplaceAll(filename, ext, "")
	)

	handler, exist := whisper.DefaultWhisperProvider.Get(input.Provider)
	if !exist {
		return fmt.Errorf("whisper provider '%s' not exist", input.Provider)
	}

	tempPath := path.Join(service.tempAudioPath, pureFilename)
	subtitleTempPath := path.Join(service.tempAudioPath, strings.ReplaceAll(filename, ext, ".ass"))

	if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
		return err
	}

	subtitleBuffer := new(bytes.Buffer)
	defer func() {
		if err := os.WriteFile(input.SavePath, subtitleBuffer.Bytes(), os.ModePerm); err != nil {
			logrus.Warnf("write target file error,fallback write to tempfile:%s", subtitleTempPath)
			_ = os.WriteFile(subtitleTempPath, subtitleBuffer.Bytes(), os.ModePerm)
		}
	}()

	subtitleSub := NewSubtitleSub()
	comment := fmt.Sprintf("generate input info: Provider:%s,Temperature:%.2f", input.Provider, input.Temperature)
	subtitleSub.subtitles.Metadata.Comments = []string{comment}
	subtitleSub.subtitles.Metadata.Title = pureFilename

	outputFileList, err := service.getAudioFromMediaWithFFmpeg(input.InputPath, tempPath, filename, handler.MaximumFileSize())
	if err != nil {
		return err
	}

	var (
		accumulateDuration = 0.0
		sequence           = int64(0)
		totalFile          = len(outputFileList)
		startTime          = time.Now() // 记录开始时间
	)

	for i, audioPath := range outputFileList {
		var (
			whisperOutput *whisper.WhisperOutput
			whisperInput  = whisper.WhisperInput{
				FilePath:    audioPath,
				Prompt:      input.Prompt,
				Temperature: input.Temperature,
				Language:    input.Language,
			}
			cacheKey = service.cacheKey(input.Provider, whisperInput)
		)

		logrus.Infof("process segment file '%s' in whisper,progress:%d/%d", audioPath, i+1, totalFile)
		if err := service.cache.Get(cacheKey, &whisperOutput); err != nil {
			return err
		} else if whisperOutput == nil {
			whisperOutput, err = handler.HandleWhisper(ctx, whisperInput)
			if err != nil {
				return err
			}
			if err := service.cache.Set(cacheKey, whisperOutput); err != nil {
				return err
			}

		}

		for _, segment := range whisperOutput.Segments {
			startTimestamp := segment.Start + accumulateDuration
			endTimestamp := segment.End + accumulateDuration
			sequence++

			segmentText := segment.Text
			segmentText = ReplaceRepeatedWords(segmentText)

			translationText := ""
			if input.TranslateTo != "" {
				translationText, err = service.translation.Translate(ctx, segmentText, "", input.TranslateTo)
				if err != nil {
					return err
				}
				translationText = ReplaceRepeatedWords(translationText)
				translationText = SegmentationByPunctuation(translationText, " ")
			}

			if translationText == "" {
				subtitleSub.AddLine(int(sequence), startTimestamp, endTimestamp, segmentText, "")
			} else {
				subtitleSub.AddLine(int(sequence), startTimestamp, endTimestamp, translationText, segmentText)
			}

		}

		accumulateDuration += whisperOutput.Duration
	}

	fileDuration := time.Duration(float64(time.Second) * accumulateDuration).String()
	processDuration := time.Since(startTime).String()
	logrus.Infof("success process file '%s' in whisper,file duration:%s,process duration:%s", input.InputPath, fileDuration, processDuration)

	if err := subtitleSub.WriteToFile(subtitleBuffer); err != nil {
		return nil
	}


	_ = os.RemoveAll(tempPath)
	

	return nil
}

func (service *SubtitleService) getAudioFromMediaWithFFmpeg(inputFile string, ouputDir string, outputName string, segmentSize int64) ([]string, error) {
	var (
		suffix            = ""
		segmentTime int64 = 0
		ext               = path.Ext(outputName)
	)

	if segmentSize > 0 {
		// 采样率 × 采样位深 × 声道数 × 时长 / 8
		segmentTime = segmentSize * 8 / 16 / 16000
		suffix = "%03d"
	}
	dirEntryList, err := os.ReadDir(ouputDir)
	if err != nil {
		return nil, err
	}

	if len(dirEntryList) == 0 {
		fileName := fmt.Sprintf("%s_%s%s", outputName[:len(outputName)-len(ext)], suffix, ".wav")
		outputPath := path.Join(ouputDir, fileName)
		if err := media.ConvertToWavWithFFmpeg(inputFile, outputPath, media.ConvertToWavOption{SegmentTime: segmentTime}); err != nil {
			return nil, err
		}

		dirEntryList, err = os.ReadDir(ouputDir)
		if err != nil {
			return nil, err
		}
	}

	fileList := make([]string, 0)

	for _, dirEntry := range dirEntryList {
		if dirEntry.IsDir() {
			continue
		}
		fileList = append(fileList, path.Join(ouputDir, dirEntry.Name()))
	}
	logrus.Infof("output file list %+v", fileList)

	return fileList, nil
}

// func (service *SubtitleService) formatTimestamp(seconds float64) string {
// 	h := int(seconds) / 3600
// 	m := (int(seconds) % 3600) / 60
// 	s := int(seconds) % 60
// 	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
// }

// func (service *SubtitleService) writeSubtitlesLine(writer io.Writer, sequence int64, startTimestamp float64, endTimestamp float64, text string) (int, error) {
// 	startTime := service.formatTimestamp(startTimestamp)
// 	endTime := service.formatTimestamp(endTimestamp)
// 	// 生成字幕行
// 	subtitleLine := fmt.Sprintf("%d\n%s --> %s\n%s\n\n", sequence, startTime, endTime, text)
// 	return writer.Write([]byte(subtitleLine))
// }

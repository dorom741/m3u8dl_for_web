package subtitle

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/model"

	"m3u8dl_for_web/pkg/media"
	"m3u8dl_for_web/pkg/split_writer"
	"m3u8dl_for_web/pkg/whisper"
	"m3u8dl_for_web/service/translation"
)

type SubtitleService struct {
	tempAudioPath string
	splitSize     int64
	translation   translation.ITranslation
	cache         *infra.FileCache
}

func NewSubtitleService(tempAudioPath string, cache *infra.FileCache, translation translation.ITranslation) *SubtitleService {
	RegisterWhisperProvider()

	return &SubtitleService{
		tempAudioPath: tempAudioPath,
		splitSize:     1024 * 1024 * 15, // 15MB
		translation:   translation,
		cache:         cache,
	}
}

func (service *SubtitleService) cacheKey(provider string, input whisper.WhisperInput) string {
	prefix := "subtitle_service"

	filename := filepath.Base(input.FilePath)
	ext := filepath.Ext(input.FilePath)

	return fmt.Sprintf("%s_%s_%s_%f", prefix, filename[:len(filename)-len(ext)], provider, input.Temperature)
}

func (service *SubtitleService) GenerateSubtitle(ctx context.Context, input model.SubtitleInput) error {
	var (
		filename = filepath.Base(input.InputPath)
		ext      = filepath.Ext(filename)
	)

	processFunc, exist := whisper.DefaultWhisperProvider.Get(input.Provider)
	if !exist {
		return fmt.Errorf("whisper provider '%s' not exist", input.Provider)
	}

	tempPath := path.Join(service.tempAudioPath, strings.ReplaceAll(filename, ext, ""))
	subtitleTempPath := path.Join(service.tempAudioPath, strings.ReplaceAll(filename, ext, ".srt"))

	if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
		return err
	}

	subtitleTempFile, err := os.OpenFile(subtitleTempPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer subtitleTempFile.Close()

	outputFileList, err := service.getAudioFromMediaWithFFmpeg(input.InputPath, tempPath, filename)
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

		if err := service.cache.Get(cacheKey, &whisperOutput); err != nil {
			return err
		} else if whisperOutput == nil {
			infra.Logger.Infof("process segment file '%s' in whisper,progress:%d/%d", audioPath, i+1, totalFile)
			whisperOutput, err = processFunc.HandleWhisper(ctx, whisperInput)
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
			}

			if _, err := service.writeSubtitlesLine(subtitleTempFile, sequence, startTimestamp, endTimestamp, fmt.Sprintf("%s\n%s", segmentText, translationText)); err != nil {
				return err
			}

			//if _, err := service.writeSubtitlesLine(subtitleFile, startTime, endTime, segment.Text); err != nil {
			//	return err
			//}
		}

		accumulateDuration += whisperOutput.Duration
	}

	infra.Logger.Infof("success process file '%s' in whisper,duration %s", input.InputPath, time.Since(startTime).String())

	_ = subtitleTempFile.Close()
	if err = os.Rename(subtitleTempPath, input.SavePath); err != nil {
		return err

	}

	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempPath)

	return nil
}

func (service *SubtitleService) getAudioFromMediaWithFFmpeg(inputFile string, ouputDir string, outputName string) ([]string, error) {
	ext := path.Ext(outputName)
	fileName := fmt.Sprintf("%s_%s%s", outputName[:len(outputName)-len(ext)], "%03d", ".wav")
	// fileName := fmt.Sprintf("%s%s", "%03d", ".wav")

	outputPath := path.Join(ouputDir, fileName)

	if err := media.ConvertToWavWithFFmpeg(inputFile, outputPath, media.ConvertToWavOption{SegmentTime: 800}); err != nil {
		return nil, err
	}

	dirEntryList, err := os.ReadDir(ouputDir)
	if err != nil {
		return nil, err
	}

	fileList := make([]string, 0)

	for _, dirEntry := range dirEntryList {
		if dirEntry.IsDir() {
			continue
		}
		fileList = append(fileList, path.Join(ouputDir, dirEntry.Name()))
	}
	infra.Logger.Infof("output file list %+v", fileList)

	return fileList, nil
}

// Deprecated
func (service *SubtitleService) getAudioFromMedia(inputFile string, ouputDir string, outputName string) ([]string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	firstBlock := make([]byte, 512)
	if _, err := file.Read(firstBlock); err != nil {
		return nil, err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	outputPath := path.Join(ouputDir, outputName)
	rotateFileWriter, err := split_writer.NewRotateFileWriter(outputPath, service.splitSize)
	if err != nil {
		return nil, err
	}
	defer rotateFileWriter.Close()

	contentType := http.DetectContentType(firstBlock)
	FirstContentType := strings.Split(contentType, "/")[0]
	switch FirstContentType {
	case "video":
		if err = media.DemuxAudio(file, rotateFileWriter); err != nil {
			return nil, err
		}
	case "audio":

		if err = media.MuxMp3ForSplit(file, rotateFileWriter); err != nil {
			return nil, err
		}
	}

	return rotateFileWriter.WritedFileList(), nil
}

func (service *SubtitleService) formatTimestamp(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (service *SubtitleService) writeSubtitlesLine(writer io.Writer, sequence int64, startTimestamp float64, endTimestamp float64, text string) (int, error) {
	startTime := service.formatTimestamp(startTimestamp)
	endTime := service.formatTimestamp(endTimestamp)
	// 生成字幕行
	subtitleLine := fmt.Sprintf("%d\n%s --> %s\n%s\n\n", sequence, startTime, endTime, text)
	return writer.Write([]byte(subtitleLine))
}

package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"m3u8dl_for_web/pkg/media"
	"m3u8dl_for_web/pkg/split_writer"
)

type SubtitleService struct {
	tempAudioPath string
	splitSize     int64
}

func NewSubtitleService(tempAudioPath string) *SubtitleService {
	return &SubtitleService{
		tempAudioPath: tempAudioPath,
		splitSize:     1024 * 1024 * 15, // 15MB
	}
}

func (service *SubtitleService) GenerateSubtitle(ctx context.Context, inputPath string, savePath string) error {
	var (
		filename           = filepath.Base(inputPath)
		ext                = filepath.Ext(filename)
		accumulateDuration = 0.0
	)

	filename = strings.ReplaceAll(filename, ext, "")
	outputFilename := strings.ReplaceAll(filename, ext, ".mp3")
	tempPath := path.Join(service.tempAudioPath, filename)

	if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
		return err
	}

	subtitleFile, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer subtitleFile.Close()

	outputFileList, err := service.extractAudioFromVideo(inputPath, tempPath, outputFilename)
	if err != nil {
		return err
	}

	for _, fileItem := range outputFileList {
		audioPath := path.Join(tempPath, fileItem)
		audioTranslationResult, err := GroqServiceInstance.AudioTranslation(ctx, audioPath)
		if err != nil {
			return err
		}

		for _, segment := range audioTranslationResult.Segments {
			startTime := segment.Start + accumulateDuration
			endTime := segment.End + accumulateDuration
			if _, err := service.writeSubtitlesLine(subtitleFile, startTime, endTime, segment.Text); err != nil {
				return err
			}
		}

		accumulateDuration += audioTranslationResult.Duration
	}

	return nil
}

func (service *SubtitleService) extractAudioFromVideo(inputFile string, ouputDir string, outputName string) ([]string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	outputPath := path.Join(ouputDir, outputName)

	rotateFileWriter, err := split_writer.NewRotateFileWriter(outputPath, service.splitSize)
	if err != nil {
		return nil, err
	}
	defer rotateFileWriter.Close()
	if err = media.SplitAudio(file, rotateFileWriter); err != nil {
		return nil, err
	}

	return rotateFileWriter.WritedFileList(), nil
}

func (service *SubtitleService) formatTimestamp(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func (service *SubtitleService) writeSubtitlesLine(writer io.Writer, startTimestamp float64, endTimestamp float64, text string) (int, error) {
	startTime := service.formatTimestamp(startTimestamp)
	endTime := service.formatTimestamp(endTimestamp)
	// 生成字幕行
	subtitleLine := fmt.Sprintf("%s --> %s\n%s\n\n", startTime, endTime, text)
	return writer.Write([]byte(subtitleLine))
}

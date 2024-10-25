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

	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/model"

	"m3u8dl_for_web/pkg/media"
	"m3u8dl_for_web/pkg/split_writer"
	"m3u8dl_for_web/pkg/whisper"
)

type SubtitleService struct {
	tempAudioPath string
	splitSize     int64
}

func NewSubtitleService(tempAudioPath string) *SubtitleService {
	RegisterWhisperProvider()

	return &SubtitleService{
		tempAudioPath: tempAudioPath,
		splitSize:     1024 * 1024 * 15, // 15MB
	}
}

func (service *SubtitleService) GenerateSubtitle(ctx context.Context, input model.SubtitleInput) error {
	var (
		filename           = filepath.Base(input.InputPath)
		ext                = filepath.Ext(filename)
		accumulateDuration = 0.0
	)

	processFunc, exist := whisper.DefaultWhisperProvider.Get(input.Provider)
	if !exist {
		return fmt.Errorf("whisper provider '%s' not exist", input.Provider)
	}

	tempPath := path.Join(service.tempAudioPath, strings.ReplaceAll(filename, ext, ""))
	print(tempPath)

	if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
		return err
	}

	subtitleFile, err := os.OpenFile(input.SavePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer subtitleFile.Close()

	outputFileList, err := service.getAudioFromMediaWithFFmpeg(input.InputPath, tempPath, filename)
	if err != nil {
		return err
	}

	for _, audioPath := range outputFileList {
		result, err := processFunc.HandleWhisper(ctx, whisper.WhisperInput{
			FilePath:    audioPath,
			Prompt:      input.Prompt,
			Temperature: input.Temperature,
			Language:    input.Language,
		})
		if err != nil {
			return err
		}

		for _, segment := range result.GetSegmentList() {
			startTime := segment.Start + accumulateDuration
			endTime := segment.End + accumulateDuration
			if _, err := service.writeSubtitlesLine(subtitleFile, startTime, endTime, segment.Text); err != nil {
				return err
			}
		}

		accumulateDuration += result.GetDuration()
	}

	return nil
}

func (service *SubtitleService) getAudioFromMediaWithFFmpeg(inputFile string, ouputDir string, outputName string) ([]string, error) {
	ext := path.Ext(outputName)
	fileName := fmt.Sprintf("%s_%s%s", outputName[:len(ext)-1], "%03d", ".wav")
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

func (service *SubtitleService) writeSubtitlesLine(writer io.Writer, startTimestamp float64, endTimestamp float64, text string) (int, error) {
	startTime := service.formatTimestamp(startTimestamp)
	endTime := service.formatTimestamp(endTimestamp)
	// 生成字幕行
	subtitleLine := fmt.Sprintf("%s --> %s\n%s\n\n", startTime, endTime, text)
	return writer.Write([]byte(subtitleLine))
}

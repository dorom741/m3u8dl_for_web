package service

import (
	"os"

	"m3u8dl_for_web/pkg/extract_audio"
	"m3u8dl_for_web/pkg/split_writer"
)

type MediaService struct{}

func (service *MediaService) SplitAudio(inputFile string, ouputDir string, outputName string) ([]string, error) {
	var splitSize int64 = 1024 * 1024 * 15

	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rotateFileWriter, err := split_writer.NewRotateFileWriter(outputName, splitSize)
	if err != nil {
		return nil, err
	}
	defer rotateFileWriter.Close()
	extract_audio.SplitAudio(file, rotateFileWriter)

	return rotateFileWriter.WritedFileList(), nil
}

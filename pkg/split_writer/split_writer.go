package split_writer

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
)

type RotateFileWriter struct {
	mu sync.Mutex

	maxSize     int64
	file        *os.File
	currentSize int64
	fileIndex   int
	fileHistory []string

	filename string
}

func NewRotateFileWriter(filename string, maxSize int64) (*RotateFileWriter, error) {
	writer := &RotateFileWriter{
		maxSize:     maxSize,
		fileIndex:   0,
		filename:    filename,
		fileHistory: make([]string, 0),
	}
	return writer, writer.createNewFile()
}

func (writer *RotateFileWriter) createNewFile() error {
	if writer.file != nil {
		writer.file.Close() // 关闭之前的文件
	}

	ext := path.Ext(writer.filename)

	fileName := strings.ReplaceAll(writer.filename, ext, fmt.Sprintf("%d%s", writer.fileIndex, ext))
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	writer.fileHistory = append(writer.fileHistory, fileName)

	writer.file = file
	writer.currentSize = 0
	writer.fileIndex++
	return nil
}

func (writer *RotateFileWriter) WritedFileList() []string {
	return writer.fileHistory
}

func (writer *RotateFileWriter) Write(p []byte) (n int, err error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	remaining := writer.maxSize - writer.currentSize
	hasWrite := 0
	if remaining == 0 {
		if err := writer.createNewFile(); err != nil {
			return 0, err
		}
		return writer.Write(p)
	}

	if remaining > int64(len(p)) {
		n, err := writer.file.Write(p)
		writer.currentSize += int64(n)
		return n, err
	}

	n, err = writer.file.Write(p[:remaining])
	if err != nil {
		return 0, err
	}
	hasWrite += n

	if err := writer.createNewFile(); err != nil {
		return 0, err
	}
	n, err = writer.Write(p[remaining:])
	return n + int(remaining), err
}

func (writer *RotateFileWriter) Close() error {
	if writer.file != nil {
		return writer.file.Close()
	}
	return nil
}

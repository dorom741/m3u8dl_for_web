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

	fileName := strings.ReplaceAll(writer.filename, ext, fmt.Sprintf("_%d%s", writer.fileIndex, ext))
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	fmt.Printf("fileName: %v\n", fileName)

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
	fmt.Printf("writer.file.Name(): %v\n", writer.file.Name())
	writer.mu.Lock()
	defer writer.mu.Unlock()

	return writer.write(p)
}

func (writer *RotateFileWriter) write(data []byte) (n int, err error) {
	remaining := writer.maxSize - writer.currentSize
	hasWrite := 0

	if remaining > int64(len(data)) {
		n, err := writer.file.Write(data)
		writer.currentSize += int64(n)
		return n, err
	}

	if remaining < int64(len(data)) {
		n, err = writer.file.Write(data[:remaining])
		if err != nil {
			return 0, err
		}
		hasWrite += n
		data = data[remaining:]
	}

	if err := writer.createNewFile(); err != nil {
		return 0, err
	}

	n, err = writer.write(data)
	return n + int(remaining), err
}

func (writer *RotateFileWriter) Close() error {
	if writer.file != nil {
		return writer.file.Close()
	}
	return nil
}

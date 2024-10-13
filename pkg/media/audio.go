package media

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ConvertMp3ToWavWithFFmpeg(inputFile string, outputPath string) error {
	// ffmpeg -i example.mp4 -vn -acodec pcm_s16le -ar 16000  -ac 1   -f segment -segment_time 600    example%03d.wav
	// 文件大小（字节） = 采样率 × 采样位深 × 声道数 × 时长 / 8
	// 900秒 ->27M
	var (
		out bytes.Buffer
	)
	ffmpegArgs := []string{
		"-i", inputFile,
		"-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1",
		"-f", "segment", "-segment_time", "900",
		outputPath,
	}

	ffmpegExecPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return err
	}

	cmd := exec.Command(ffmpegExecPath, ffmpegArgs...)
	cmd.Stdout = &out
	cmd.Stderr = &out

	fmt.Printf("exec ffmpeg command line:%s", cmd.String())

	err = cmd.Run()

	return err

}

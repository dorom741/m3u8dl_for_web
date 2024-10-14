package media

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
)

type ConvertToWavOption struct {
	// Default 16000
	SampleRate int64

	// Default 1
	Channel int64

	//Set zero to no segment,unit seconds
	SegmentTime int64

	OverrideOnExist bool
}

func ConvertToWavWithFFmpeg(inputFile string, outputPath string, option ConvertToWavOption) error {
	// ffmpeg -i example.mp4 -vn -acodec pcm_s16le -ar 16000  -ac 1   -f segment -segment_time 600    example%03d.wav
	// 文件大小（字节） = 采样率 × 采样位深 × 声道数 × 时长 / 8
	// 900秒 ->27M
	var (
		out bytes.Buffer
		overrideOnExistFlag = "-n"
	)

	if option.Channel <= 0 {
		option.Channel = 1
	}

	if option.SampleRate <= 0 {
		option.SampleRate = 16000
	}

	if option.OverrideOnExist {
		overrideOnExistFlag = "-y"
	}

	ffmpegArgs := []string{
		"-hide_banner", "-i", inputFile,overrideOnExistFlag,
		"-vn", "-acodec", "pcm_s16le", "-ar", strconv.FormatInt(option.SampleRate, 10), "-ac", strconv.FormatInt(option.Channel, 10),
	}

	if option.SegmentTime > 0 {
		ffmpegArgs = append(ffmpegArgs, "-f", "segment", "-segment_time", strconv.FormatInt(option.SegmentTime, 10))
	}

	ffmpegArgs = append(ffmpegArgs, outputPath)

	ffmpegExecPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return err
	}

	cmd := exec.Command(ffmpegExecPath, ffmpegArgs...)
	cmd.Stdout = &out
	cmd.Stderr = &out

	fmt.Printf("exec ffmpeg command line:%s\n", cmd.String())

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("exec ffmpeg err command line:'%s' error:%s", cmd.String(), out.String())
	}

	return nil

}

package media

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-mp4"
	"github.com/yapingcat/gomedia/go-mpeg2"
)

func DemuxAudio(input io.ReadSeeker, ouput io.Writer) error {
	demuxer := mp4.CreateMp4Demuxer(input)
	if infos, err := demuxer.ReadHead(); err != nil && err != io.EOF {
		return err
	} else {
		logrus.Debugf("media infos %+v", infos)
	}
	mp4info := demuxer.GetMp4Info()
	logrus.Debugf("media info %+v\n", mp4info)
	for {
		pkg, err := demuxer.ReadPacket()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		logrus.Debugf("track:%d,cid:%+v,pts:%d dts:%d\n", pkg.TrackId, pkg.Cid, pkg.Pts, pkg.Dts)
		if pkg.Cid == mp4.MP4_CODEC_AAC || pkg.Cid == mp4.MP4_CODEC_MP3 {
			if _, err := ouput.Write(pkg.Data); err != nil {
				return err
			}
		}
	}

	return nil
}

func MuxMp3ForSplit(input io.Reader, ouput io.Writer) error {
	muxer := mpeg2.NewTSMuxer()
	muxer.OnPacket = func(pkg []byte) {
		ouput.Write(pkg)
	}

	pid := muxer.AddStream(mpeg2.TS_STREAM_AUDIO_MPEG1)
	mp3, err := io.ReadAll(input)
	if err != nil {
		return err
	}
	var pts uint64 = 0
	var dts uint64 = 0
	var codecErr error
	codec.SplitMp3Frames(mp3, func(head *codec.MP3FrameHead, frame []byte) {
		sampleSize := head.SampleSize
		sampleRate := head.GetSampleRate()
		fmt.Println("sampleSize", sampleSize, "sampleRate:", sampleRate)

		delta := sampleSize * 1000 / sampleRate
		if err := muxer.Write(pid, frame, pts, dts); err != nil {
			codecErr = err
			return
		}
		fmt.Println("write pts:", pts, "dts:", dts)
		pts += uint64(delta)
		dts = pts
	})

	return codecErr
}

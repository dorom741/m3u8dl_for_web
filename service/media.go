package service

import (
	"fmt"
	"io"

	"github.com/yapingcat/gomedia/go-mp4"
)

type MediaService struct{}

func (service *MediaService) SplitAudio(input io.ReadSeeker, ouput io.Writer) error {
	demuxer := mp4.CreateMp4Demuxer(input)
	if infos, err := demuxer.ReadHead(); err != nil && err != io.EOF {
		return err
	} else {
		fmt.Printf("infos %+v\n", infos)
	}
	mp4info := demuxer.GetMp4Info()
	fmt.Printf("%+v\n", mp4info)
	for {
		pkg, err := demuxer.ReadPacket()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fmt.Printf("track:%d,cid:%+v,pts:%d dts:%d\n", pkg.TrackId, pkg.Cid, pkg.Pts, pkg.Dts)
		if pkg.Cid == mp4.MP4_CODEC_AAC || pkg.Cid == mp4.MP4_CODEC_MP3 {
			ouput.Write(pkg.Data)
		}
	}

	return nil
}

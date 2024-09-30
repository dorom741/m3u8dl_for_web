package extract_audio

import (
	"io"

	"github.com/yapingcat/gomedia/go-mp4"

	"m3u8dl_for_web/infra"
)

func  SplitAudio(input io.ReadSeeker, ouput io.Writer) error {
	demuxer := mp4.CreateMp4Demuxer(input)
	if infos, err := demuxer.ReadHead(); err != nil && err != io.EOF {
		return err
	} else {
		infra.Logger.Debugf("media infos %+v", infos)
	}
	mp4info := demuxer.GetMp4Info()
	infra.Logger.Debugf("media info %+v\n", mp4info)
	for {
		pkg, err := demuxer.ReadPacket()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		infra.Logger.Debugf("track:%d,cid:%+v,pts:%d dts:%d\n", pkg.TrackId, pkg.Cid, pkg.Pts, pkg.Dts)
		if pkg.Cid == mp4.MP4_CODEC_AAC || pkg.Cid == mp4.MP4_CODEC_MP3 {
			if _, err := ouput.Write(pkg.Data); err != nil {
				return err
			}
		}
	}

	return nil
}
package extract_audio

import (
	"os"
	"testing"
)




var (
	// ctx := context.Background()
	inputPath = "/sdcard/Download/meeting_01.mp4"
	// inputPath := "/data/data/com.termux/files/home/workplace/project/m3u8dl_for_web/resource/download/meeting_01_0.mp4"
	tempPath = "/data/data/com.termux/files/home/workplace/project/m3u8dl_for_web/resource/download/meeting_temp.mp4"
	outputPath = "/data/data/com.termux/files/home/workplace/project/m3u8dl_for_web/resource/download/meeting_test.mp3"

)


func TestSplitAudio(t *testing.T) {

	f, err := os.Open(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()


	
	tempFile, err := os.Create(tempPath)
	if err != nil {
		t.Fatal(err)

	}
	defer tempFile.Close()



	if err := SplitAudio(f, tempFile); err != nil {
		t.Fatal(err)

	}

}




func TestMuxMp3ForSplit(t *testing.T) {
	tempFile, err := os.Open(tempPath)
	if err != nil {
		t.Fatal(err)

	}
	defer tempFile.Close()


	tsf, err := os.Create(outputPath)
	if err != nil {
		t.Fatal(err)

	}
	defer tsf.Close()


	if err := MuxMp3ForSplit(tempFile, tsf); err != nil {
		t.Fatal(err)

	}

}

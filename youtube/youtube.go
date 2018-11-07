package youtube

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type stream map[string]string

type Youtube struct {
	url         string
	quality     string
	downloadDir string
	extractMp3  bool
	videoId     string
	videoInfo   string
	streamList  []stream
}

const baseVideoInfoUrl string = "http://youtube.com/get_video_info?video_id="

func YoutubeObj(url string, quality string, downloadDir string, extractMp3 bool) *Youtube {
	return &Youtube{url: url, quality: quality, downloadDir: downloadDir, extractMp3: extractMp3}
}

func (y *Youtube) Download() error {
	if err := y.findVideoId(); err != nil {
		return fmt.Errorf("Video Id not found: %s", err)
	}
	if err := y.fetchVideoInfo(); err != nil {
		return fmt.Errorf("Video info error: %s", err)
	}
	if err := y.DecodeVideoInfo(); err != nil {
		return fmt.Errorf("Video decode error: %s", err)
	}
	var streamIndex int
	switch y.quality {
	default:
	case "high":
		streamIndex = 0
	case "low":
		streamIndex = len(y.streamList) - 1
	case "medium":
		streamIndex = (len(y.streamList) - 1) / 2
	}
	stream := y.streamList[streamIndex]
	ext := y.findExtension(stream["type"])
	fileName := fmt.Sprintf("%s.%s", stream["title"], ext)
	log.Printf("Downloading file: %s\n", fileName)
	err := y.StartDownload(fileName, stream["url"])
	if err == nil {
		log.Println("Download Completed.")
	}
	if y.extractMp3 {
		err = y.extractAsMp3(fileName)
		if err != nil {
			log.Println(err)
		}
	}
	return err
}

func (y *Youtube) findExtension(videoType string) string {
	re := regexp.MustCompile("video/(\\w+);")
	if re.MatchString(videoType) {
		groups := re.FindStringSubmatch(videoType)
		return groups[1]
	}
	return "mp4"
}

func (y *Youtube) findVideoId() error {
	videoID := y.url
	if strings.Contains(videoID, "youtu") || strings.ContainsAny(videoID, "\"?&/<%=") {
		reList := []*regexp.Regexp{
			regexp.MustCompile(`(?:v|embed|watch\?v)(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`(?:=|/)([^"&?/=%]{11})`),
			regexp.MustCompile(`([^"&?/=%]{11})`),
		}
		for _, re := range reList {
			if isMatch := re.MatchString(videoID); isMatch {
				subs := re.FindStringSubmatch(videoID)
				videoID = subs[1]
			}
		}
	}
	log.Printf("Found video id: '%s'", videoID)
	y.videoId = videoID
	if strings.ContainsAny(videoID, "?&/<%=") {
		return errors.New("invalid characters in video id")
	}
	if len(videoID) < 10 {
		return errors.New("the video id must be at least 10 characters long")
	}
	return nil
}

func (y *Youtube) fetchVideoInfo() error {
	url := baseVideoInfoUrl + y.videoId
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("Video info status code is not 200")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	y.videoInfo = string(body)
	return nil
}

func (y *Youtube) DecodeVideoInfo() error {
	qResponse, err := url.ParseQuery(y.videoInfo)
	if err != nil {
		return err
	}
	status, ok := qResponse["status"]
	if !ok {
		return errors.New("status not found in video info")
	}
	if status[0] != "ok" {
		return errors.New("Server response status is not ok")
	}
	streamMap, ok := qResponse["url_encoded_fmt_stream_map"]
	if !ok {
		return errors.New("No stream map found")
	}
	streamList := strings.Split(streamMap[0], ",")
	var streams []stream
	for streamIndex, rawStream := range streamList {
		qStream, err := url.ParseQuery(rawStream)
		if err != nil {
			log.Printf("Vide stream cannot be parsed: Index: %d: %s", streamIndex, err)
			continue
		}
		var singleStream stream = make(stream)
		for key, value := range qStream {
			singleStream[key] = value[0]
		}
		singleStream["title"] = qResponse["title"][0]
		singleStream["author"] = qResponse["author"][0]
		streams = append(streams, singleStream)
	}
	y.streamList = streams
	return nil
}

func (y *Youtube) StartDownload(fileName string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("Status code is not 200")
	}
	err = os.MkdirAll(filepath.Dir(y.downloadDir), 666)
	if err != nil {
		return err
	}
	out, err := os.Create(y.downloadDir + "/" + fileName)
	if err != nil {
		return err
	}
	mw := io.MultiWriter(out)
	_, err = io.Copy(mw, resp.Body)
	if err != nil {
		log.Println("download video err=", err)
		return err
	}
	return nil
}

func (y *Youtube) extractAsMp3(fileName string) error {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return err
	}
	mp3 := strings.TrimRight(fileName, filepath.Ext(fileName)) + ".mp3"
	cmd := exec.Command(ffmpeg, "-y", "-loglevel", "quiet", "-i", y.downloadDir+"/"+fileName, "-vn", y.downloadDir+"/"+mp3)
	err = cmd.Run()
	if err == nil {
		log.Println("Mp3 extracted: ", mp3)
	}
	return err
}

func (y *Youtube) GetUrl() string {
	return y.url
}

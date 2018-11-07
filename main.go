package main

import (
	"flag"
	"fmt"
	"os"
	"youtubedl/youtube"
)

func main() {
	var url string
	var quality string
	var dir string
	var mp3 string
	flag.StringVar(&url, "url", "", "Enter youtube url")
	flag.StringVar(&quality, "quality", "high", "Quality: low|medium|high")
	flag.StringVar(&mp3, "mp3", "no", "Extract Mp3: yes|no")

	currentDir, _ := os.Getwd()
	flag.StringVar(&dir, "dir", currentDir, "Enter download directory")
	flag.Parse()
	if dir == "" {
		fmt.Println("Download directory not found!")
		return
	}
	if url == "" {
		flag.Usage()
		return
	}
	extractMp3 := false
	if mp3 == "yes" {
		extractMp3 = true
	}
	y := youtube.YoutubeObj(url, quality, dir, extractMp3)
	y.Download()
}

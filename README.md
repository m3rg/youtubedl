# Youtube Video Downloader in Golang
Download youtube videos using binary or the library file in golang. Convert video files to mp3 if ffmpeg is intalled.
### Installation
```
go get https://github.com/m3rg/youtubedl
```
### Usage of Binary
```
./youtubedl -url="https://www.youtube.com/watch?v=89O3Kh8rUrg"
```
Parameters
```
  -url string
        Enter youtube url or video ID
  -dir string
        Enter download directory (default: working directory)
  -mp3 string
        Extract Mp3: yes|no (default "no")
  -quality string
        Quality: low|medium|high (default "high")
```
### Usage of Library
```go
package main

import (
	"os"

	"github.com/m3rg/youtubedl/youtube"
)

func main() {
	url := "89O3Kh8rUrg" // Url or video ID
	quality := "low"
	currentDir, _ := os.Getwd()
	extractMp3 := false
	y := youtube.YoutubeObj(url, quality, currentDir, extractMp3)
	y.Download()
}
```

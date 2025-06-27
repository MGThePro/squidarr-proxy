package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

var DownloadPath string
var Category string
var Port string
var Region string
var ApiLink string
var ApiKey string

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func main() {
	DownloadPath = getEnv("DOWNLOAD_PATH", "/data/squidarr/")
	Category = getEnv("CATEGORY", "music")
	Region = getEnv("REGION", "eu")
	Port = getEnv("PORT", "8687")
	ApiLink = "https://" + Region + ".qobuz.squid.wtf/api"
	ApiKey = getEnv("API_KEY", "")

	//create folders if they don't exist yet
	os.Mkdir(DownloadPath, 0775)
	os.Mkdir(DownloadPath+"/incomplete", 0775)
	os.Mkdir(DownloadPath+"/incomplete/"+Category, 0775)
	os.Mkdir(DownloadPath+"/complete", 0775)
	os.Mkdir(DownloadPath+"/complete/"+Category, 0775)

	//and now clear anything in /incomplete that was created by squidarr. Likely a leftover failed download
	folders, err := os.ReadDir(DownloadPath + "/incomplete/" + Category)
	if err != nil {
		fmt.Println("Couldn't read incomplete folder: ")
		fmt.Println(err)
	}
	for _, folder := range folders {
		if strings.Contains(folder.Name(), "-SQUIDWTF") {
			fmt.Println("Removing incomplete download " + folder.Name())
			err := os.RemoveAll(DownloadPath + "/incomplete/" + Category + "/" + folder.Name())
			if err != nil {
				fmt.Println("Failed to remove folder!")
				fmt.Println(err)
			}
		}
	}

	// Generate a basic list of downloads from folders in /complete. likely from completed downloads that weren't imported before reboot.
	// Adding these to the downloads list allows importing/deleting from Lidarr
	folders, _ = os.ReadDir(DownloadPath + "/complete/" + Category)
	for _, folder := range folders {
		if strings.Contains(folder.Name(), "-SQUIDWTF") {
			fmt.Println("Adding completed download " + folder.Name() + " to history")
			var download Download
			download.FileName = folder.Name()
			//Don't really care about this anymore, but making sure they're equal so they show up in the history, not the queue
			download.numTracks = 1
			download.downloaded = 1
			//Can't know the exact ID anymore, but all it's needed for now is as a NZO_ID so generating a random one...
			b := make([]byte, 13)
			for i := range b {
				b[i] = byte(rand.Intn(27) + 65)
			}
			download.Id = string(b)
			Downloads[download.Id] = &download
		}
	}

	http.HandleFunc("/indexer", handleIndexerRequest)
	http.HandleFunc("/downloader/api", handleDownloaderRequest)
	fmt.Println("Listening on port " + Port + "...")
	http.ListenAndServe(":"+Port, nil)
}

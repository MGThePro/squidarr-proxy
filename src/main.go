package main

import (
	"fmt"
	"net/http"
	"os"
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

	http.HandleFunc("/indexer", handleIndexerRequest)
	http.HandleFunc("/downloader/api", handleDownloaderRequest)
	fmt.Println("Listening on port " + Port + "...")
	http.ListenAndServe(":"+Port, nil)
}

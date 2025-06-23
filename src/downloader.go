package main

import (
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type File struct {
	Id           int
	Name         string
	DownloadLink string
	completed    bool
}

type Download struct {
	Id         string
	Artist     string
	numTracks  int
	downloaded int
	FileName   string
	Files      []File
}

var Downloads map[string]*Download = make(map[string]*Download)

func handleDownloaderRequest(w http.ResponseWriter, r *http.Request) {
	switch query := r.URL.Query().Get("mode"); query {
	case "get_config":
		get_config(w, *r.URL)
	case "version":
		version(w, *r.URL)
	case "addfile":
		addfile(w, r)
	case "queue":
		queue(w, r)
	case "history":
		history(w, r)
	default:
		fmt.Println("Downloader unknown request:")
		fmt.Println(r.Method)
		fmt.Println(r.URL.String())
		fmt.Println(r.Header)
		buffer := make([]byte, 100)
		for {
			n, err := r.Body.Read(buffer)
			fmt.Printf("%q\n", buffer[:n])
			if err == io.EOF {
				break
			}
		}
		w.Write([]byte("Request received!"))
	}
}

func get_config(w http.ResponseWriter, u url.URL) {
	w.Write([]byte(`{
	    "config": {
	        "misc": {
	            "complete_dir": "/data/complete",
	            "enable_tv_sorting": false,
	            "enable_movie_sorting": false,
	            "pre_check": false,
	            "history_retention": "",
	            "history_retention_option": "all"
	        },
	        "categories": [
	            {
	                "name": "music",
	                "pp": "",
	                "script": "Default",
	                "dir": "/data/complete/music",
	                "priority": -100
	            },
	        ],
	        "sorters": []
	    }
	}`))
}

func version(w http.ResponseWriter, u url.URL) {
	w.Write([]byte(`{
 	    "version": "4.5.1"
 	}`))
}

func addfile(w http.ResponseWriter, r *http.Request) {
	//extract filename, QobuzId and number of tracks
	var body []byte = make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil && err != io.EOF {
		fmt.Println("/downloader/api/addfile Failed to read body:")
		fmt.Println(err)
	}
	reNum := regexp.MustCompile("[a-zA-Z0-9]+")
	reName := regexp.MustCompile("filename=.*.nzb")
	var lines []string = strings.Split(string(body), "\n")
	var filename string = reName.FindString(lines[1])
	filename = strings.Trim(filename, "filename=\"")
	filename = strings.TrimRight(filename, ".nzb")
	var Id = reNum.FindString(lines[6])
	fmt.Println(filename)
	var NumTracks, _ = strconv.Atoi(reNum.FindString(lines[7]))
	generateDownload(filename, Id, NumTracks)
	//send response using QobuzId as nzo_id
	w.Write([]byte("{\n" +
		"\"status\": true,\n" +
		"\"nzo_ids\": [\"SABnzbd_nzo_" + Id + "\"]\n" +
		"}"))
}

func generateDownload(filename string, Id string, numTracks int) {
	var download Download
	download.Id = Id
	download.numTracks = numTracks
	download.FileName = filename
	download.downloaded = 0
	var queryUrl string = ApiLink + "/get-album?album_id=" + Id
	resp, err := http.Get(queryUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	//making the request body usable
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	download.Artist = gjson.Get(string(bodyBytes), "data.artist.name").String()
	result := gjson.Get(string(bodyBytes), "data.tracks.items")
	result.ForEach(func(key, value gjson.Result) bool {
		var track File
		track.Id = int(gjson.Get(value.String(), "id").Int())
		track.Name = gjson.Get(value.String(), "title").String()
		track.completed = false
		var queryUrl string = ApiLink + "/download-music?track_id=" + strconv.Itoa(track.Id) + "&quality=27"
		resp, err := http.Get(queryUrl)
		if err != nil {
			fmt.Println(err)
			return false
		}
		//making the request body usable
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return false
		}
		track.DownloadLink = gjson.Get(string(bodyBytes), "data.url").String()
		download.Files = append(download.Files, track)
		return true
	})
	Downloads[Id] = &download
	go startDownload(&download)
}

func queue(w http.ResponseWriter, r *http.Request) {
	var response string = "{\n" +
		"	\"queue\": {\n" +
		"		\"paused\": false,\n" +
		"		\"slots\": ["

	//fill slots with current download queue
	var index int = 0
	for id := range Downloads {
		var download Download = *Downloads[id]
		if download.downloaded == download.numTracks {
			//shouldnt be in queue anymore, skipping
			break
		}
		//Don't know how long the download will take, so estimating 10 seconds per track remaining
		timeleft := (download.numTracks - download.downloaded) * 10
		//Guessing progress based on how many tracks are left, not based on file size
		progress := (int((float64(download.downloaded) / float64(download.numTracks)) * 100))
		response += "\n{\n" +
			"			\"status\": \"Downloading\",\n" +
			"			\"index\": " + strconv.Itoa(index) + ",\n" +
			//mostly answering the same garbage, hope Lidarr doesn't pay attention...
			"			\"password\": \"\",\n" +
			"			\"avg_age\": \"2895d\",\n" +
			"			\"script\": \"None\",\n" +
			"			\"direct_unpack\": \"30/30\",\n" +
			//claiming every download is 100mb so mbleft is just 100-progress
			"			\"mb\": \"" + "100" + "\",\n" +
			"			\"mbleft\": \"" + strconv.Itoa(100-progress) + "\",\n" +
			"			\"mbmissing\": \"0.0\",\n" +
			"			\"size\": \"100 MB\",\n" +
			"			\"sizeleft\": \"" + strconv.Itoa(100-progress) + " MB\",\n" +
			"			\"filename\": \"" + download.FileName + "\",\n" +
			"			\"labels\": [],\n" +
			"			\"priority\": \"Normal\",\n" +
			"			\"cat\": \"" + Category + "\",\n" +
			"			\"timeleft\": \"0:" + strconv.Itoa(timeleft/60) + ":" + strconv.Itoa(timeleft%60) + "\",\n" +
			"			\"percentage\": \"" + strconv.Itoa(progress) + "\",\n" +
			"			\"nzo_id\": \"SABnzbd_nzo_" + download.Id + "\",\n" +
			"			\"unpackopts\": \"3\"\n" +
			"},\n"
	}

	response += "]\n" +
		"	}\n" +
		"}"
	w.Write([]byte(response))
}

func history(w http.ResponseWriter, r *http.Request) {
	var response string = `{
	    "history": {
	        "slots": [`
	//fill this with completed history
	for id := range Downloads {
		var download Download = *Downloads[id]
		if download.downloaded < download.numTracks {
			//not finished yet, skipping...
			break
		}
		// Get the fileinfo
		fileInfo, _ := os.Stat(DownloadCompletePath + "/" + Category + "/" + download.FileName)
		response += "\n{\n" +
			"\"name\": \"" + download.FileName + "\", \n" +
			"\"nzb_name\": \"" + download.FileName + ".nzb\",\n" +
			"\"category\": \"" + Category + "\",\n" +
			"\"bytes\": " + strconv.FormatInt(fileInfo.Size(), 10) + ",\n" +
			//same estimate of 10 seconds per track, could measure time in the future
			"\"download_time\": " + strconv.Itoa(download.numTracks*30) + ",\n" +
			"\"status\": \"Completed\",\n" +
			"\"storage\": \"" + DownloadCompletePath + "/" + Category + "/" + download.FileName + "\",\n" +
			"\"nzo_id\": \"SABnzbd_nzo_" + download.Id + "\"\n" +
			"},"
	}
	response += `]
	    }
	}`
	w.Write([]byte(response))
}

func startDownload(download *Download) {
	//create folder
	err := os.Mkdir(DownloadIncompletePath+"/"+Category+"/"+download.FileName, 0755)
	if err != nil {
		fmt.Println("Couldn't create folder in " + DownloadIncompletePath + "/" + Category)
		fmt.Println(err)
	}
	//Download each track
	for _, track := range download.Files {
		var Path string = DownloadIncompletePath + "/" + Category + "/" + download.FileName + "/" + download.Artist + " - " + track.Name + ".flac"
		_, err := grab.Get(Path, track.DownloadLink)
		if err != nil {
			fmt.Println("Failed to download track " + track.Name)
			fmt.Println(err)
		} else {
			track.completed = true
			download.downloaded += 1
		}
	}
	//Download (should be) complete, move to complete folder
	os.Rename(DownloadIncompletePath+"/"+Category+"/"+download.FileName, DownloadCompletePath+"/"+Category+"/"+download.FileName)
}

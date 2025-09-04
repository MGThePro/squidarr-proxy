package main

import (
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"github.com/tidwall/gjson"
	"go.senan.xyz/taglib"
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
	Index        string
	mediaNumber  string
	isrc         string
	DownloadLink string
	completed    bool
}

type Download struct {
	Id         string
	Artist     string
	Album      string
	Comment    string
	CoverUrl   string
	numTracks  int
	mediaCount int
	label      string
	downloaded int
	FileName   string
	Files      []File
}

var Downloads map[string]*Download = make(map[string]*Download)

func handleDownloaderRequest(w http.ResponseWriter, r *http.Request) {
	var queryApiKey string = r.URL.Query().Get("apikey")
	if ApiKey != queryApiKey {
		w.Write([]byte("error: API Key Incorrect"))
		return
	}
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
	            "complete_dir": "` + DownloadPath + `/complete",
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
	                "dir": "` + DownloadPath + `/incomplete/music",
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
	if Downloads[Id].downloaded != -1 {
		go startDownload(Id)
	}
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
	download.Album = gjson.Get(string(bodyBytes), "data.title").String()
	download.Comment = gjson.Get(string(bodyBytes), "data.version").String()
	download.CoverUrl = gjson.Get(string(bodyBytes), "data.image.large").String()
	if download.Comment == "null" {
		download.Comment = ""
	}
	download.numTracks = int(gjson.Get(string(bodyBytes), "data.tracks_count").Int())
	download.mediaCount = int(gjson.Get(string(bodyBytes), "data.media_count").Int())
	download.label = gjson.Get(string(bodyBytes), "data.copyright").String()
	result := gjson.Get(string(bodyBytes), "data.tracks.items")
	result.ForEach(func(key, value gjson.Result) bool {
		var track File
		var valueString = value.String()
		track.Id = int(gjson.Get(valueString, "id").Int())
		track.Name = gjson.Get(valueString, "title").String()
		track.Index = gjson.Get(valueString, "track_number").String()
		track.mediaNumber = gjson.Get(valueString, "media_number").String()
		track.isrc = gjson.Get(valueString, "isrc").String()
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
		if track.DownloadLink == "" {
			fmt.Println("squid.wtf didn't give a link for track " + track.Name)
			fmt.Println("This is most likely an error with squid.wtf, Qobuz-DL or Qobuz itself...")
			fmt.Println("Cancelling download...")
			download.downloaded = -1
			err := os.RemoveAll(DownloadPath + "/incomplete/" + Category + "/" + download.FileName)
			if err != nil {
				fmt.Println("Couldn't delete folder " + download.FileName)
				fmt.Println(err)
			}
			return false
		}
		download.Files = append(download.Files, track)
		return true
	})
	Downloads[Id] = &download
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
	//check for deletion call first
	//api?mode=history&name=delete&del_files=1&value=SABnzbd_nzo_0825646642830&archive=1&apikey=(removed)&output=json
	if r.URL.Query().Get("name") == "delete" {
		var id, _ = strings.CutPrefix(r.URL.Query().Get("value"), "SABnzbd_nzo_")
		if r.URL.Query().Get("del_files") == "1" {
			err := os.RemoveAll(DownloadPath + "/complete/" + Category + "/" + Downloads[id].FileName)
			if err != nil {
				fmt.Println("Couldn't delete folder " + Downloads[id].FileName)
				fmt.Println(err)
			}
		}
		delete(Downloads, id)
	}
	var response string = `{
	    "history": {
	        "slots": [`
	//fill this with completed history
	for id := range Downloads {
		var download Download = *Downloads[id]
		if download.downloaded < download.numTracks && download.downloaded != -1 {
			//not finished yet, skipping...
			break
		}
		// Get the fileinfo
		fileInfo, err := os.Stat(DownloadPath + "/complete/" + Category + "/" + download.FileName)
		var fileSize int64
		if err != nil {
			//cant get file stats on Docker for some reason? giving arbitrary size info
			fileSize = 10000
		} else {
			fileSize = fileInfo.Size()
		}
		var status string
		if download.downloaded == -1 {
			status = "Failed"
		} else {
			status = "Completed"
		}
		response += "\n{\n" +
			"\"name\": \"" + download.FileName + "\", \n" +
			"\"nzb_name\": \"" + download.FileName + ".nzb\",\n" +
			"\"category\": \"" + Category + "\",\n" +
			"\"bytes\": " + strconv.FormatInt(fileSize, 10) + ",\n" +
			//same estimate of 10 seconds per track, could measure time in the future
			"\"download_time\": " + strconv.Itoa(download.numTracks*30) + ",\n" +
			"\"status\": \"" + status + "\",\n" +
			"\"storage\": \"" + DownloadPath + "/complete/" + Category + "/" + download.FileName + "\",\n" +
			"\"nzo_id\": \"SABnzbd_nzo_" + download.Id + "\"\n" +
			"},"
	}
	response += `]
	    }
	}`
	w.Write([]byte(response))
}

func startDownload(Id string) {
	download := Downloads[Id]
	//create folder
	var Folder string = DownloadPath + "/incomplete/" + Category + "/" + download.FileName + "/"
	err := os.Mkdir(Folder, 0755)
	if err != nil {
		fmt.Println("Couldn't create folder in " + DownloadPath + "/incomplete/" + Category)
		fmt.Println(err)
		return
	}
	//Download cover art
	_, err = grab.Get(Folder+"cover.jpg", download.CoverUrl)
	if err != nil {
		fmt.Println("Failed to download cover")
		fmt.Println(err)
		return
	}
	//Download each track
	for _, track := range download.Files {
		var Name string = track.Index + " - " + download.Artist + " - " + track.Name + ".flac"
		_, err := grab.Get(Folder+Name, track.DownloadLink)
		if err != nil {
			fmt.Println("Failed to download track " + track.Name)
			fmt.Println(err)
			return
		} else {
			track.completed = true
			download.downloaded += 1
			writeMetaData(*download, track, Folder+Name)
		}
	}
	//Download (should be) complete, move to complete folder
	os.Rename(Folder, DownloadPath+"/complete/"+Category+"/"+download.FileName)
}

func writeMetaData(album Download, track File, fileName string) {
	err := taglib.WriteTags(fileName, map[string][]string{
		taglib.AlbumArtist: {album.Artist},
		taglib.Artist:      {album.Artist},
		taglib.Album:       {album.Album},
		taglib.TrackNumber: {track.Index},
		taglib.Title:       {track.Name},
		taglib.Comment:     {album.Comment},
		taglib.DiscNumber:  {track.mediaNumber},
		taglib.Label:       {album.label},
		taglib.ISRC:        {track.isrc},
	}, 0)
	if err != nil {
		fmt.Println("Couldn't write Metadata to file " + fileName)
		fmt.Println(err)
	}
}

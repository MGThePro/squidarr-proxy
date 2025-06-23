package main

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Album struct {
	Artist       string
	Title        string
	Edition      string
	ReleaseDate  int64
	Publisher    string
	CoverUrl     string
	SamplingRate int64
	BitDepth     int64
	Id           string
	NumTracks    int64
	Channels     int64
	Duration     int64
	Size         int64
}

func handleIndexerRequest(w http.ResponseWriter, r *http.Request) {
	switch query := r.URL.Query().Get("t"); query {
	case "caps":
		caps(w, *r.URL)
	case "music":
		music(w, *r.URL)
	case "search":
		search(w, *r.URL)
	case "fakenzb":
		fakenzb(w, *r.URL)
	default:
		fmt.Println("Indexer unknown request:")
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

func caps(w http.ResponseWriter, u url.URL) {
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<caps>
    <limits max="5000" default="5000"/>
    <registration available="no" open="no"/>
    <searching>
        <search available="yes" supportedParams="q"/>
        <tv-search available="no" supportedParams=""/>
        <movie-search available="no" supportedParams=""/>
        <audio-search available="yes" supportedParams="q" />
    </searching>
    <categories>
        <category id="3000" name="Audio">
            <subcat id="3010" name="Audio/MP3"/>
            <subcat id="3020" name="Audio/Video"/>
            <subcat id="3030" name="Audio/Audiobook"/>
            <subcat id="3040" name="Audio/Lossless"/>
            <subcat id="3050" name="Audio/Podcast"/>
        </category>
    </categories>
</caps>
	`))
}

func music(w http.ResponseWriter, u url.URL) {
	if u.Query().Get("q") == "" {
		fmt.Println("searching with no query, responding garbage...")
		w.Write([]byte(
			`<?xml version="1.0" encoding="UTF-8"?>
<rss xmlns:newznab="http://www.newznab.com/DTD/2010/feeds/attributes/" version="2.0">
  <channel>
    <title>example.com</title>
    <description>example.com API results</description>
    <newznab:response offset="0" total="1234"/>

    <item>
      <!-- Standard RSS 2.0 data -->
      <title>A.Public.Domain.Album.Name</title>
      <guid isPermaLink="true">http://servername.com/rss/viewnzb/e9c515e02346086e3a477a5436d7bc8c</guid>
      <link>http://servername.com/rss/nzb/e9c515e02346086e3a477a5436d7bc8c&amp;i=1&amp;r=18cf9f0a736041465e3bd521d00a90b9</link>
      <comments>http://servername.com/rss/viewnzb/e9c515e02346086e3a477a5436d7bc8c#comments</comments>
      <pubDate>Sun, 06 Jun 2010 17:29:23 +0100</pubDate>
      <category>Music > MP3</category>
      <description>Some music</description>
      <enclosure url="http://servername.com/rss/nzb/e9c515e02346086e3a477a5436d7bc8c&amp;i=1&amp;r=18cf9f0a736041465e3bd521d00a90b9" length="154653309" type="application/x-nzb" />

      <!-- Additional attributes -->
      <newznab:attr name="category" value="3000" />
      <newznab:attr name="category" value="3010" />
      <newznab:attr name="size"     value="144967295" />
      <newznab:attr name="artist"   value="Bob Smith" />
      <newznab:attr name="album"    value="Groovy Tunes" />
      <newznab:attr name="publisher" value="Epic Music" />
      <newznab:attr name="year"     value="2011" />
      <newznab:attr name="tracks"   value="track one|track two|track three" />
      <newznab:attr name="coverurl" value="http://servername.com/covers/music/12345.jpg" />
      <newznab:attr name="review"   value="This album is great" />
    </item>

  </channel>
</rss>
		`))
		return
	}
}

func search(w http.ResponseWriter, u url.URL) {
	//doing the actual querying request
	//getting the query parameters
	limit, err := strconv.Atoi(u.Query().Get("limit"))
	if err != nil {
		limit = 10
	}
	var query string = strings.Replace(u.Query().Get("q"), " ", "+", -1)
	offset, err := strconv.Atoi(u.Query().Get("offset"))
	if err != nil {
		offset = 0
	}
	var Albums []Album
	//squid.wtf seems to only be able to output 10 items (albums/tracks each) at once
	//so iterate over 10 items at a time until reaching the limit...
	for i := 0; i < limit; i += 10 {
		var queryUrl string = ApiLink + "/get-music?q=" + query + "&offset=" + (strconv.Itoa(offset + i)) + "&limit=10"
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
		//check the number of results and modify limit to avoid unnecessary requests
		//lidarr seems to always start with 100
		if limit == 100 {
			total := int(gjson.Get(string(bodyBytes), "data.albums.total").Int())
			if total < 100 {
				limit = total
			}
		}
		//iterate over each album and create an Album struct object from it
		result := gjson.Get(string(bodyBytes), "data.albums.items")
		result.ForEach(func(key, value gjson.Result) bool {
			var album Album
			var resultString string = value.String()
			album.Artist = gjson.Get(resultString, "artist.name").String()
			album.Title = gjson.Get(resultString, "title").String()
			album.Edition = gjson.Get(resultString, "version").String()
			album.ReleaseDate = gjson.Get(resultString, "released_at").Int()
			album.Publisher = gjson.Get(resultString, "label.name").String()
			album.CoverUrl = gjson.Get(resultString, "image.small").String()
			album.SamplingRate = gjson.Get(resultString, "maximum_sampling_rate").Int()
			album.BitDepth = gjson.Get(resultString, "maximum_bit_depth").Int()
			album.Id = gjson.Get(resultString, "id").String()
			album.NumTracks = gjson.Get(resultString, "tracks_count").Int()
			album.Channels = gjson.Get(resultString, "maximum_channel_count").Int()
			album.Duration = gjson.Get(resultString, "duration").Int()
			//guesstimate filesize based on Sampling Rate, Bit Depth, Channel count and duration
			//assuming all tracks of that album have the same specifications and that FLAC is 70% as large as WAV
			// (Sampling Rate in Hz * Bit depth * channels * seconds) / 8 to get it from bits to bytes
			album.Size = int64(float64(((album.SamplingRate * 1000) * (album.BitDepth * album.Channels * album.Duration) / 8)) * 0.7)
			Albums = append(Albums, album)
			return true // keep iterating
		})
	}
	//Create XML Response
	var response string = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<rss xmlns:newznab=\"http://www.newznab.com/DTD/2010/feeds/attributes/\" version=\"2.0\">\n" +
		"<channel>\n" +
		"<title>example.com</title>\n" +
		"<description>example.com API results</description>\n" +
		"<newznab:apilimits apicurrent=\"123\" apimax=\"500\" grabcurrent=\"69\" grabmax=\"250\" apioldesttime=\"Wed, 17 Jul 2019 23:00:49 +0100\" graboldesttime=\"Thu, 18 Jul 2019 04:44:44 +0100\" />\n" +
		"    <newznab:response offset=\"0\" total=\"" + strconv.Itoa(len(Albums)) + "\"/>`)\n"

	//iterate over each album and create <item> parts of response
	for _, album := range Albums {
		//some basic sanitation of artist and title
		reg := regexp.MustCompile("[^a-zA-Z0-9 ]+")
		album.Title = reg.ReplaceAllString(album.Title, "")
		album.Artist = reg.ReplaceAllString(album.Artist, "")

		response += "<item>" +
			"    <!-- Standard RSS 2.0 Data -->" +
			"    <title>" + releaseName(album) + "</title>" +
			"    <guid isPermaLink=\"true\">http://servername.com/rss/viewnzb/e9c515e02346086e3a477a5436d7bc8c</guid>" +
			"    <link>http://servername.com/rss/nzb/e9c515e02346086e3a477a5436d7bc8c&amp;i=1&amp;r=18cf9f0a736041465e3bd521d00a90b9</link>" +
			"    <comments>http://servername.com/rss/viewnzb/e9c515e02346086e3a477a5436d7bc8c#comments</comments>" +
			"    <pubDate>" + time.Unix(album.ReleaseDate, 0).Format("Mon, 02 Jan 2006 15:04:05 -0700") + "</pubDate>" +
			"    <category>Audio > Lossless</category>" +
			"    <description>" + album.Artist + " " + album.Title + "</description>" +
			"    <enclosure url=\"/indexer?t=fakenzb&amp;qobuzid=" + album.Id + "&amp;numtracks=" + strconv.FormatInt(album.NumTracks, 10) + "\" length=\"" + strconv.FormatInt(album.Size, 10) + "\" type=\"application/x-nzb\" />" +

			"    <!-- Additional attributes -->" +
			"    <newznab:attr name=\"category\" value=\"3000\"/>" +
			"    <newznab:attr name=\"category\" value=\"3040\"/>" +
			"    <newznab:attr name=\"size\"     value=\"" + strconv.FormatInt(album.Size, 10) + "\"/>" +
			"    </item>"
	}

	//write the end of the response
	response += "</channel>\n" +
		"</rss>"

	w.Write([]byte(response))
}

func releaseName(album Album) (name string) {
	release := time.Unix(album.ReleaseDate, 0)
	name = album.Artist + "-" + album.Title + " (" + album.Edition + ")" + "-" + strconv.FormatInt(album.BitDepth, 10) + "BIT-" + strconv.FormatInt(album.SamplingRate, 10) + "-KHZ-WEB-FLAC-" + strconv.Itoa(release.Year()) + "-SQUIDWTF"
	return name
}

func fakenzb(w http.ResponseWriter, u url.URL) {
	QobuzId := u.Query().Get("qobuzid")
	NumTracks := u.Query().Get("numtracks")
	w.Header().Set("Content-Type", "application/x-nzb")
	response := "<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n" +
		"<!DOCTYPE nzb PUBLIC \"-//newzBin//DTD NZB 1.0//EN\" \"http://www.newzbin.com/DTD/nzb/nzb-1.0.dtd\">\n" +
		"<!-- " + QobuzId + "  -->\n" +
		"<!-- " + NumTracks + " -->\n" +
		"<nzb>\n" +
		"    <file post_id=\"1\">\n" +
		"        <groups>\n" +
		"            <group>squidwtf</group>\n" +
		"        </groups>\n" +
		"        <segments>\n" +
		"            <segment number=\"1\">ExampleSegmentID@news.example.com</segment>\n" +
		"        </segments>\n" +
		"    </file>\n" +
		"</nzb>"
	w.Write([]byte(response))
}

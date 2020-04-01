package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type metadataResponse struct {
	MediaContainer *struct {
		Size                int    `json:"size"`
		AllowSync           bool   `json:"allowSync"`
		AugmentationKey     string `json:"augmentationKey"`
		Identifier          string `json:"identifier"`
		LibrarySectionID    int    `json:"librarySectionID"`
		LibrarySectionTitle string `json:"librarySectionTitle"`
		LibrarySectionUUID  string `json:"librarySectionUUID"`
		MediaTagPrefix      string `json:"mediaTagPrefix"`
		MediaTagVersion     int    `json:"mediaTagVersion"`
		Metadata            []*struct {
			RatingKey             string   `json:"ratingKey"`
			Key                   string   `json:"key"`
			ParentRatingKey       string   `json:"parentRatingKey"`
			GrandparentRatingKey  string   `json:"grandparentRatingKey"`
			GUID                  string   `json:"guid"`
			ParentGUID            string   `json:"parentGuid"`
			GrandparentGUID       string   `json:"grandparentGuid"`
			Type                  string   `json:"type"`
			Title                 string   `json:"title"`
			GrandparentKey        string   `json:"grandparentKey"`
			ParentKey             string   `json:"parentKey"`
			LibrarySectionTitle   string   `json:"librarySectionTitle"`
			LibrarySectionID      int      `json:"librarySectionID"`
			LibrarySectionKey     string   `json:"librarySectionKey"`
			GrandparentTitle      string   `json:"grandparentTitle"`
			ParentTitle           string   `json:"parentTitle"`
			ContentRating         string   `json:"contentRating"`
			Summary               string   `json:"summary"`
			Index                 int      `json:"index"`
			ParentIndex           int      `json:"parentIndex"`
			Rating                float64  `json:"rating"`
			Year                  int      `json:"year"`
			Thumb                 string   `json:"thumb"`
			Art                   string   `json:"art"`
			ParentThumb           string   `json:"parentThumb"`
			GrandparentThumb      string   `json:"grandparentThumb"`
			GrandparentArt        string   `json:"grandparentArt"`
			GrandparentTheme      string   `json:"grandparentTheme"`
			Duration              int      `json:"duration"`
			OriginallyAvailableAt string   `json:"originallyAvailableAt"`
			AddedAt               int      `json:"addedAt"`
			UpdatedAt             int      `json:"updatedAt"`
			Media                 []*media `json:"Media"`
			Writer                []*struct {
				ID     int    `json:"id"`
				Filter string `json:"filter"`
				Tag    string `json:"tag"`
			} `json:"Writer"`
			Extras *struct {
				Size int `json:"size"`
			} `json:"Extras"`
		} `json:"Metadata"`
	} `json:"MediaContainer"`
}

type media struct {
	Title           string  `json:"title"`
	ID              int     `json:"id"`
	Duration        int     `json:"duration"`
	Bitrate         int     `json:"bitrate"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	AspectRatio     float64 `json:"aspectRatio"`
	AudioChannels   int     `json:"audioChannels"`
	AudioCodec      string  `json:"audioCodec"`
	VideoCodec      string  `json:"videoCodec"`
	VideoResolution string  `json:"videoResolution"`
	Container       string  `json:"container"`
	VideoFrameRate  string  `json:"videoFrameRate"`
	VideoProfile    string  `json:"videoProfile"`
	Part            []*struct {
		ID           int    `json:"id"`
		Key          string `json:"key"`
		Duration     int    `json:"duration"`
		File         string `json:"file"`
		Size         int    `json:"size"`
		Container    string `json:"container"`
		VideoProfile string `json:"videoProfile"`
		Stream       []*struct {
			ID                 int     `json:"id"`
			StreamType         int     `json:"streamType"`
			Default            bool    `json:"default"`
			Codec              string  `json:"codec"`
			Index              int     `json:"index"`
			Bitrate            int     `json:"bitrate"`
			Language           string  `json:"language"`
			LanguageCode       string  `json:"languageCode"`
			BitDepth           int     `json:"bitDepth,omitempty"`
			ChromaLocation     string  `json:"chromaLocation,omitempty"`
			ChromaSubsampling  string  `json:"chromaSubsampling,omitempty"`
			CodedHeight        string  `json:"codedHeight,omitempty"`
			CodedWidth         string  `json:"codedWidth,omitempty"`
			ColorPrimaries     string  `json:"colorPrimaries,omitempty"`
			ColorRange         string  `json:"colorRange,omitempty"`
			ColorSpace         string  `json:"colorSpace,omitempty"`
			ColorTrc           string  `json:"colorTrc,omitempty"`
			FrameRate          float64 `json:"frameRate,omitempty"`
			HasScalingMatrix   bool    `json:"hasScalingMatrix,omitempty"`
			Height             int     `json:"height,omitempty"`
			Level              int     `json:"level,omitempty"`
			Profile            string  `json:"profile,omitempty"`
			RefFrames          int     `json:"refFrames,omitempty"`
			ScanType           string  `json:"scanType,omitempty"`
			Width              int     `json:"width,omitempty"`
			DisplayTitle       string  `json:"displayTitle"`
			Selected           bool    `json:"selected,omitempty"`
			Channels           int     `json:"channels,omitempty"`
			AudioChannelLayout string  `json:"audioChannelLayout,omitempty"`
			SamplingRate       int     `json:"samplingRate,omitempty"`
			Title              string  `json:"title,omitempty"`
		} `json:"Stream"`
	} `json:"Part"`
}

var r = regexp.MustCompile(".* \\((.*)\\)$")

func getMediaTitle(media *media) string {
	for _, part := range media.Part {
		if part.Stream == nil {
			continue
		}

		for _, stream := range part.Stream {
			if stream.StreamType == 1 && stream.DisplayTitle != "" {
				format := r.FindStringSubmatch(stream.DisplayTitle)

				if format[1] != "" {
					return format[1] + " " + media.Container
				}

				return media.Container
			}
		}
	}

	return "Unknown"
}

func main() {
	bindAddress := flag.String("addr", "127.0.0.1:3213", "the address to bind to")
	plexHost := flag.String("plex-host", "localhost:32401", "the host + port that your plex server is running on")
	secure := flag.Bool("secure", false, "use https to connect to your plex server (will increase loading times) (needed if Secure Connections is set to Required)")

	flag.Parse()

	log.Printf("binding to %s\n", *bindAddress)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.String(), "/library/metadata/") {
			return
		}

		url := r.URL

		url.Host = *plexHost

		if *secure {
			url.Scheme = "http"
		} else {
			url.Scheme = "https"
		}

		request, _ := http.NewRequest(r.Method, url.String(), nil)

		for key, values := range r.Header {
			if key == "Accept-Encoding" {
				continue
			}

			for _, value := range values {
				request.Header.Add(key, value)
			}
		}

		response, err := http.DefaultClient.Do(request)
		if err != nil {
			log.Fatalf("error while firing request to server: %s\n", err.Error())
			return
		}
		defer response.Body.Close()

		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("error while reading response body: %s\n", err.Error())
			return
		}

		var parsed metadataResponse

		err = json.Unmarshal(data, &parsed)
		if err != nil {
			log.Printf("error while parsing response: %s\n", err.Error())
			w.Write(data)
			return
		}

		if parsed.MediaContainer == nil || parsed.MediaContainer.Metadata == nil {
			w.Write(data)
			return
		}

		for _, meta := range parsed.MediaContainer.Metadata {
			if meta.Media == nil {
				continue
			}

			for _, media := range meta.Media {
				if media.Part == nil {
					continue
				}

				if media.Title == "" {
					media.Title = getMediaTitle(media)
				}

				for _, part := range media.Part {
					if part.Stream == nil {
						continue
					}

					for _, stream := range part.Stream {
						if stream.DisplayTitle == "" || stream.Title == "" {
							continue
						}

						stream.DisplayTitle = stream.DisplayTitle + " (" + stream.Title + ")"
					}
				}
			}
		}

		data, err = json.Marshal(parsed)
		if err != nil {
			log.Printf("error while marshaling response: %s\n", err.Error())
			w.Write(data)
			return
		}

		for key, values := range response.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.Write(data)
	})

	err := http.ListenAndServe(*bindAddress, nil)
	if err != nil {
		log.Fatalf("error while starting server: %s\n", err.Error())
		return
	}
}

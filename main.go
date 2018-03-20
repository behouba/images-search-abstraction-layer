package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// QwantImageAPI represent the json data get from the qwant api call
type QwantImageAPI struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}

type Data struct {
	Result Result `json:"result"`
}

type Result struct {
	Total int    `json:"total"`
	Items []Item `json:"items"`
}

type Item struct {
	URL       string `json:"media"`
	Snipet    string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	Context   string `json:"url"`
}

// The Qwant api endpoint
const endPoint = "https://api.qwant.com/api/search/images?count=20&"

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/", homePage)
	router.HandleFunc("/api/imagesearch/", imageSearch)

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	server.ListenAndServe()
}

// home page to display index.html template
func homePage(w http.ResponseWriter, r *http.Request) {
	templ := template.Must(template.ParseFiles("template/index.html"))
	templ.Execute(w, nil)
}

func imageSearch(w http.ResponseWriter, r *http.Request) {
	var searchInput string
	var offset int

	// trying to get string to make image search with the qwant api
	searchInput = r.URL.Path[17:]
	// check if the path matched
	if strings.Count(searchInput, "/") > 0 {
		fmt.Fprint(w, "CAN NOT GET: "+r.URL.Path)
		return
	}
	// trying to get the value of offset query if it is specified
	offsetSlice := r.URL.Query()["offset"]
	if offsetSlice != nil {
		var err error
		// try to convert offset query value to integer
		offset, err = strconv.Atoi(offsetSlice[0])
		if err != nil {
			fmt.Println("Error converting offset value into integer: ", err)
			return
		}
	}
	// get jsonData to display based on searchInput and offset value
	jsonData, err := getImages(searchInput, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprint(w, jsonData)
}

func getImages(search string, offset int) (string, error) {
	// GET request to the qwant api with the specified offset and searchInput
	resp, err := http.Get(endPoint + "offset=" + url.QueryEscape(strconv.Itoa(offset)) + "&q=" + url.QueryEscape(search))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	// trying to read the body of the data get from the api call
	jsonDataByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	var respData QwantImageAPI
	// Umnarshal data from api call into local struct
	json.Unmarshal(jsonDataByte, &respData)

	// get the values needed from the entire struct
	ImagesData := respData.Data.Result.Items

	// make json data from ImagesData
	ImageJSONData, err := json.Marshal(ImagesData)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(ImageJSONData), nil
}

package main

import (
	"encoding/json"
	"errors"
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

// Data struct from QwantImageAPI
type Data struct {
	Result Result `json:"result"`
}

// Result struct from QwantImageAPI
type Result struct {
	Items []Item `json:"items"`
}

// Item struct to make json data of images
type Item struct {
	URL       string `json:"media"`
	Snipet    string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	Context   string `json:"url"`
}

// Search struct representing json data of a search
type Search struct {
	Term string `json:"term"`
	When string `json:"when"`
}

// The Qwant api endpoint
const endPoint = "https://api.qwant.com/api/search/images?count=20&"

func main() {
	router := http.NewServeMux()

	files := http.FileServer(http.Dir("assets"))
	router.Handle("/assets/", http.StripPrefix("/assets/", files))

	router.HandleFunc("/", homePage)
	router.HandleFunc("/api/imagesearch/", imageSearch)
	router.HandleFunc("/api/latest/imagesearch/", latestSearch)

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
			fmt.Fprint(w, err)
			return
		}
	}

	// get jsonData to display based on searchInput and offset value
	jsonData, err := getImages(searchInput, offset)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = saveSearch(searchInput)
	if err != nil {
		fmt.Println("Failed to save data to the database :", err)
		return
	}
	fmt.Fprint(w, jsonData)
}

func getImages(search string, offset int) (string, error) {
	// GET request to the qwant api with the specified offset and searchInput
	apiCallURL := endPoint + "offset=" + url.QueryEscape(strconv.Itoa(offset)) + "&q=" + url.QueryEscape(search)
	resp, err := http.Get(apiCallURL)
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

	// check the api call status
	if respData.Status == "error" {
		return "", errors.New("API call failed try again later")
	}
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

func latestSearch(w http.ResponseWriter, r *http.Request) {
	// call the getLatestSearch function to get latest search data
	latestSearch, err := getLatestSearch()
	if err != nil {
		fmt.Fprint(w, "Error occur when trying to get latest search: ", err)
		return
	}
	// trying to make json byte data from data of latest search list
	latestSearchJSON, err := json.Marshal(latestSearch)
	if err != nil {
		fmt.Fprint(w, "Error occur when trying to marshal Search list into JSON", err)
		return
	}
	fmt.Fprint(w, string(latestSearchJSON))
}

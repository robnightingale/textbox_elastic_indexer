package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	//"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/machinebox/sdk-go/textbox"
)

type CsvLine struct {
	Column1 string
	Column2 string
}

var (
	client *http.Client
	tbox   *textbox.Client
)

func main() {
	var dataset, es, textboxAddr string
	flag.StringVar(&dataset, "dataset", "./bbcsport/bbc-text.csv", "Full path and filename of dataset")
	flag.StringVar(&es, "es", "http://localhost:9200", "Elastic Search address")
	flag.StringVar(&textboxAddr, "textbox", "http://localhost:8000", "Textbox address")
	flag.Parse()

	client = &http.Client{
		Timeout: 10 * time.Second,
	}

	tbox = textbox.New(textboxAddr)
	index := "news_textbox/articles"

	log.Println("[INFO]: Using ES on ", es)
	log.Println("[INFO]: Using Textbox on ", textboxAddr)
	log.Println("[INFO]: Start indexing articles from ", dataset, "to the index/type", index)

	insertWithTextboxES(es, index, dataset)
	log.Println("[INFO]: Finished")
}

func sendToTextBox(title string, text string) map[string]interface{} {

	resp, err := tbox.Check(strings.NewReader(text))
	if err != nil {
		return map[string]interface{}{
			"textbox": err.Error()}
	}

	keywords := []string{}
	people := []string{}
	places := []string{}
	for _, k := range resp.Keywords {
		keywords = append(keywords, k.Keyword)
	}
	for _, s := range resp.Sentences {
		for _, ent := range s.Entities {
			if ent.Type == "person" {
				people = append(people, ent.Text)
			}
			if ent.Type == "place" {
				places = append(places, ent.Text)
			}
		}
	}
	// split to get the title and the content
	body := map[string]interface{}{
		"title":    title,
		"content":  text,
		"keywords": keywords,
		"people":   people,
		"places":   places,
	}
	return body
}

// inserts an article pre-processing it with textbox on Elastic Search with this structure
// {
//	 id: "xxxxxx",
//   title: "title of the article"
//   content: "content of the article",
//   keywords: "<most relevant keywords>",
//   people: "<people named in the content>"
//   places: "<places named in the content>"
// }
func insertWithTextboxES(es, index, path string) error {
	// Open CSV file
	log.Println("Opening dataset")
	r, err := os.Open(path)
	if err != nil {
		return err
	}
	defer r.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(r).ReadAll()
	if err != nil {
		panic(err)
	}

	// Loop through lines & turn into object
	for _, line := range lines {
		data := CsvLine{
			Column1: line[0],
			Column2: line[1],
		}
		text := sendToTextBox(data.Column1, data.Column2)
		//log.Println("NLP returned ", text["title"])

		postES(es, index, path, text)
	}

	return nil
}

// Post to Elastic Search, and returns an error in case is not success
// (there are good elastic search Go libs to do that, but is easy enoughs)
func postES(es string, index string, path string, body map[string]interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	u, err := url.Parse(es)
	if err != nil {
		return err
	}
	u.Path = index
	reader := bytes.NewReader(b)
	r, err := http.NewRequest(http.MethodPost, u.String(), reader)
	r.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}

	resp, err := client.Do(r)
	if err != nil {
		return errors.New("ES error:" + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("[ERROR]: reading ES response", path, err)
		} else {
			log.Println("[ERROR]: ES error ", string(respBody))
		}
		return errors.New("Error creating article on Elastic Search " + resp.Status)
	}

	return nil
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
)

func main() {
	req, _ := http.NewRequest("POST", "http://localhost:8080/upload", nil)
	req.Header.Set("Accept", "multipart/form-data; charset=utf-8")
	resp, _ := http.DefaultClient.Do(req)
	mediaType, params, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	fmt.Println(mediaType)
	mr := multipart.NewReader(resp.Body, params["boundary"])
	for part, err := mr.NextPart(); err == nil; part, err = mr.NextPart() {
		value, _ := ioutil.ReadAll(part)
		log.Printf("Value: %s", value)
	}
}

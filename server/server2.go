// https://github.com/golang-samples/http/blob/master/fileupload/main.go
// https://peter.bourgon.org/blog/2019/02/12/multipart-http-responses.html
package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

// 1MB
const MAX_MEMORY = 1 * 1024 * 1024

func upload(w http.ResponseWriter, r *http.Request) {

	// 1. get and manage files and datas
	if err := r.ParseMultipartForm(MAX_MEMORY); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	for key, value := range r.MultipartForm.Value {
		fmt.Fprintf(w, "%s:%s ", key, value)
		log.Printf("%s:%s", key, value)
	}

	path, _ := os.Getwd()
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, _ := fileHeader.Open()
			// path := fmt.Sprintf("files/%s", fileHeader.Filename)
			filePath := path + "/" + fileHeader.Filename
			buf, _ := ioutil.ReadAll(file)
			ioutil.WriteFile(filePath, buf, os.ModePerm)
		}
	}

	// 2. send back server side files and datas
	writer := multipart.NewWriter(w)
	w.Header().Set("Content-Type", writer.FormDataContentType())

	path1 := path + "/testserver1.txt"
	path2 := path + "/testserver2.txt"

	files := map[string]string {
		"file1": path1,
		"file2": path2,
	}

	randomServerNumber := "0987654321"
	hmac := []byte{0x99,0x88,0x77,0x0e,0xc5,0xd3,0x35,0xd4,0x72,0x4f,0x3e,0xcf,0xfa,0x76,0x6a,0x4c,0x3e,0x07,0x60,0x06,0xa1,0x2b,0x78,0x75,0x52,0x0f,0x80,0x9d,0xde,0x17,0xbf,0xde}  
	// 9988770ec5d335d4724f3ecffa766a4c3e076006a12b7875520f809dde17bfde
	hmacString := hex.EncodeToString(hmac)

	datas := map[string]string {
		"data1":      randomServerNumber,  // randomNumber
		"data2":      hmacString,  // hmac
	}

	// 2-1. files
	for paramName, filePath := range files {
		// 1) open file from the path
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 2) get file as []byte AND get file info
		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		file.Close()

		// 3) create form file part1. 
		part, err := writer.CreateFormFile(paramName, fileInfo.Name())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		part.Write(fileContents)
	}

	// 2-2. add key, value
	for key, val := range datas {
		_ = writer.WriteField(key, val)
	}
	err := writer.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func main() {
	http.HandleFunc("/upload", upload)
	http.Handle("/", http.FileServer(http.Dir("static")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
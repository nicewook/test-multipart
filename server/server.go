package main

import (
	// "bytes"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"log"
	// "mime"
	"mime/multipart"
	"net/http"
	"os"
)

const MAX_MEMORY = 1 * 1024 * 1024  //  1MB

func upload(w http.ResponseWriter, r *http.Request) {

	// 1. get and manage uploaded files and data
	// https://github.com/golang-samples/http/blob/master/fileupload/main.go
	if err := r.ParseMultipartForm(MAX_MEMORY); err != nil {
		log.Println("ParseMultipartForm err:", err)
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	// parse key/value data
	rcvData := make(map[string]string)
	for key, value := range r.MultipartForm.Value {
		log.Printf("key:value %s:%s", key, value)
		log.Printf("value type: %T, %T", key, value)
		rcvData[key] = value[0]
	}

	clientRandomNumber := rcvData["data1"]
	clientHMAC, _ := base64.StdEncoding.DecodeString(rcvData["data2"])
	log.Println("clientRandomNumber:", clientRandomNumber)  // 1122334455
	log.Println("clientHMAC:", hex.EncodeToString(clientHMAC))  // 8813d40ec5d335d4724f3ecffa766a4c3e076006a12b7875520f809dde17bfde


	// parse and save files
	path, _ := os.Getwd()
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, _ := fileHeader.Open()
			filePath := path + "/" + fileHeader.Filename
			buf, _ := ioutil.ReadAll(file)
			ioutil.WriteFile(filePath, buf, os.ModePerm)
		}
	}

	// 2. response with server side files and data
	// https://peter.bourgon.org/blog/2019/02/12/multipart-http-responses.html
	mw := multipart.NewWriter(w)
	w.Header().Set("Content-Type", mw.FormDataContentType())

	// prepare files and data
	path1 := path + "/testserver1.txt"
	path2 := path + "/testserver2.txt"

	files := map[string]string {
		"file1": path1,
		"file2": path2,
	}

	randomServerNumber := "9988776655"
	hmac := []byte{0x99,0x99,0x99,0x0e,0xc5,0xd3,0x35,0xd4,0x72,0x4f,0x3e,0xcf,0xfa,0x76,0x6a,0x4c,0x3e,0x07,0x60,0x06,0xa1,0x2b,0x78,0x75,0x52,0x0f,0x80,0x9d,0xde,0x17,0xbf,0xde}  
	// 9999990ec5d335d4724f3ecffa766a4c3e076006a12b7875520f809dde17bfde
	hmacBase64 := base64.StdEncoding.EncodeToString(hmac)
	
	sndData := map[string]string {
		"data1": randomServerNumber, // randomNumber
		"data2": hmacBase64,         // hmac
	}

	// files
	for paramName, filePath := range files {
		// 1) open file from the path
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		// 2) get file as []byte AND get file info
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		file.Close()

		// 3) create form file part1. 
		part, err := mw.CreateFormFile(paramName, fileInfo.Name())
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		part.Write(fileBytes)
	}
	
	// add key, value
	for key, val := range sndData {
		_ = mw.WriteField(key, val)
	}

	// and close
	mw.Close()
}

func main() {
	http.HandleFunc("/upload", upload)
	// http.Handle("/", http.FileServer(http.Dir("static")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
package main

import (
	"bytes"
	"crypto/hmac"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

const MAX_MEMORY = 1 * 1024 * 1024  // 1MB

// Creates a new files and data upload http request with optional extra params
// https://matt.aimonetti.net/posts/2013-07-golang-multipart-file-upload-example/
func multipartRequest(uri string, files map[string]string, data map[string]string) (*http.Request, error) {

	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)

	// 1. files
	for paramName, filePath := range files {
		// 1. open file from the path
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		// 2. get file as []byte AND get file info
		fileContents, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}
		file.Close()

		// 3. create form file part1.
		part, err := mw.CreateFormFile(paramName, fileInfo.Name())
		if err != nil {
			return nil, err
		}
		part.Write(fileContents)
	}

  // 2. add key, value
	for key, val := range data {
		_ = mw.WriteField(key, val)
	}

	err := mw.Close()
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", uri, body)
	req.Header.Set("Accept", "multipart/form-data; charset=utf-8")
	req.Header.Add("Content-Type", mw.FormDataContentType())

	return req, nil
}

func main() {
	// 1. prepare files and data
	path1 := "client/test1.txt"
	path2 := "client/test2.txt"

	files := map[string]string {
		"file1": path1,
		"file2": path2,
	}

	clientRandomNumber := "1122334455"
	clientHMAC := []byte{0x88,0x13,0xd4,0x0e,0xc5,0xd3,0x35,0xd4,0x72,0x4f,0x3e,0xcf,0xfa,0x76,0x6a,0x4c,0x3e,0x07,0x60,0x06,0xa1,0x2b,0x78,0x75,0x52,0x0f,0x80,0x9d,0xde,0x17,0xbf,0xde}
	// 8813d40ec5d335d4724f3ecffa766a4c3e076006a12b7875520f809dde17bfde
	hmacBase64 := base64.StdEncoding.EncodeToString(clientHMAC)

	data := map[string]string {
		"data1": clientRandomNumber, // clientRandomNumber
		"data2": hmacBase64,  			 // hmac
	}
	request, err := multipartRequest("http://127.0.0.1:8080/upload", files, data)

	if err != nil {
    log.Fatalf("Unable to run multipartRequest: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(request)

	// https://stackoverflow.com/questions/53215506/no-output-after-multipart-newreader
	mediaType, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
			log.Fatal(err)
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(resp.Body, params["boundary"])
		form, err := mr.ReadForm(MAX_MEMORY)
		if err != nil {
			fmt.Println(err)
		}

		// parse key/value data
		rcvData := make(map[string]string)
		for key, value := range form.Value {
			// log.Printf("key:value %s:%s", key, value)
			rcvData[key] = value[0]
		}

		serverRandomNumber := rcvData["data1"]
		serverHMAC, _ := base64.StdEncoding.DecodeString(rcvData["data2"])
		log.Println("serverRandomNumber:", serverRandomNumber)  // 9988776655
		log.Println("serverHMAC:", hex.EncodeToString(serverHMAC))  // 9999990ec5d335d4724f3ecffa766a4c3e076006a12b7875520f809dde17bfde

		expectedHMAC := []byte{0x99,0x99,0x99,0x0e,0xc5,0xd3,0x35,0xd4,0x72,0x4f,0x3e,0xcf,0xfa,0x76,0x6a,0x4c,0x3e,0x07,0x60,0x06,0xa1,0x2b,0x78,0x75,0x52,0x0f,0x80,0x9d,0xde,0x17,0xbf,0xde}
		if hmac.Equal(serverHMAC, expectedHMAC) {
			log.Println("Yeah, I've got expectedHMAC!")
		} else {
			log.Println("Oh, No. Wrong HMAC from the server!")
		}

		// files
		for _, fileHeaders := range form.File {
			for _, fileHeader := range fileHeaders {
				file, _ := fileHeader.Open()
				filePath := fileHeader.Filename
				buf, _ := ioutil.ReadAll(file)
				ioutil.WriteFile(filePath, buf, os.ModePerm)
			}
		}
	}
}

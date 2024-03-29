package main

import (
	"bytes"
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

const MAX_MEMORY = 1 * 1024 * 1024

// https://gist.github.com/mattetti/5914158/f4d1393d83ebedc682a3c8e7bdc6b49670083b84
// https://matt.aimonetti.net/posts/2013-07-golang-multipart-file-upload-example/
// Creates a new file upload http request with optional extra params
func multipartRequest(uri string, files map[string]string, data map[string]string) (*http.Request, error) {

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

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
		part, err := writer.CreateFormFile(paramName, fileInfo.Name())
		if err != nil {
			return nil, err
		}
		part.Write(fileContents)
	}

  // 2. add key, value
	for key, val := range data {
		_ = writer.WriteField(key, val)
	}
	
	err := writer.Close()
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", uri, body)
	req.Header.Set("Accept", "multipart/form-data; charset=utf-8")
	req.Header.Add("Content-Type", writer.FormDataContentType())
	fmt.Println("req.Header", req.Header)

	
	// req.Header.Set("Accept", writer.FormDataContentType())
	
	return req, nil
}

func main() {
	path, _ := os.Getwd()
	path1 := path + "/test1.txt"
	path2 := path + "/test2.txt"

	files := map[string]string {
		"file1": path1,
		"file2": path2,
	}

	randomNumber := "1234567890"
	hmac := []byte{0x88,0x13,0xd4,0x0e,0xc5,0xd3,0x35,0xd4,0x72,0x4f,0x3e,0xcf,0xfa,0x76,0x6a,0x4c,0x3e,0x07,0x60,0x06,0xa1,0x2b,0x78,0x75,0x52,0x0f,0x80,0x9d,0xde,0x17,0xbf,0xde}  
	// 8813d40ec5d335d4724f3ecffa766a4c3e076006a12b7875520f809dde17bfde

	hmacString := hex.EncodeToString(hmac)

	data := map[string]string {
		"data1":       randomNumber,  // randomNumber
		"data2":      hmacString,  // hmac
	}
	request, err := multipartRequest("http://127.0.0.1:8080/upload", files, data)
	if err != nil {
		log.Fatal("fatal!", err)
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	// if err != nil {
	// 	log.Fatal("fatal2!", err)
	// } else {
	// 	var bodyContent []byte
	// 	fmt.Println(resp.StatusCode)
	// 	fmt.Println(resp.Header)
	// 	resp.Body.Read(bodyContent)
	// 	resp.Body.Close()
	// 	fmt.Println(bodyContent)
	// }

	

	// https://stackoverflow.com/questions/53215506/no-output-after-multipart-newreader
	mediaType, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
			log.Fatal(err)
	}
	fmt.Println("mediaType", mediaType)
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(resp.Body, params["boundary"])
		
		form, err := mr.ReadForm(MAX_MEMORY)
		if err != nil {
			fmt.Println(err)
		}

		for key, value := range form.Value {
			// fmt.Fprintf(w, "%s:%s ", key, value)
			log.Printf("%s:%v", key, value)
		}

		for _, fileHeaders := range form.File {
			for _, fileHeader := range fileHeaders {
				file, _ := fileHeader.Open()
				// path := fmt.Sprintf("files/%s", fileHeader.Filename)
				filePath := path + "/" + fileHeader.Filename
				fmt.Println("filePath",filePath)
				buf, _ := ioutil.ReadAll(file)
				ioutil.WriteFile(filePath, buf, os.ModePerm)
			}
		}

	}	
}
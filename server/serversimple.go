package main

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/hex"
	"io"
	// "io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

const MAX_MEMORY = 1 * 1024 * 1024  //  1MB

func getFileFromRequest(r *http.Request, fieldName string) (string, error) {
	path, _ := os.Getwd()
	path = path + "/"

	file, header, err := r.FormFile(fieldName)
	defer file.Close()
	if err != nil {
		return "", err
	}

	fileToSave, err := os.OpenFile(header.Filename, os.O_CREATE | os.O_RDWR, os.FileMode(666))
	if err != nil {
		return "", err
	}
	defer fileToSave.Close()
	io.Copy(fileToSave, file)

	return path + header.Filename, nil
}

func writeFileToMultipartWriter(mw *multipart.Writer, fieldName, fileName string) error {
	path, _ := os.Getwd()
	path = path + "/"
		
	// 1) open file from the path
	file, err := os.Open(path + fileName)
	defer file.Close()
	if err != nil {
		return err
	}	
	
	// 2) create form file part1. 
	fileForm, err := mw.CreateFormFile(fieldName, fileName)
	if err != nil {
		return err
	}

	if _, err := io.Copy(fileForm, file); err != nil {
		return err
	}	
	return nil
}

func upload(w http.ResponseWriter, r *http.Request) {

	// 1. parse from the Request
	// parse key/value data	
	clientRandomNumber := r.FormValue("data1")
	clientHMAC, _ := base64.StdEncoding.DecodeString(r.FormValue("data2"))
	log.Println("clientRandomNumber:", clientRandomNumber)  // 1122334455
	log.Println("clientHMAC:", hex.EncodeToString(clientHMAC))  // 8813d40ec5d335d4724f3ecffa766a4c3e076006a12b7875520f809dde17bfde

	// verify
	expectedHMAC := []byte{0x88,0x13,0xd4,0x0e,0xc5,0xd3,0x35,0xd4,0x72,0x4f,0x3e,0xcf,0xfa,0x76,0x6a,0x4c,0x3e,0x07,0x60,0x06,0xa1,0x2b,0x78,0x75,0x52,0x0f,0x80,0x9d,0xde,0x17,0xbf,0xde}
	if hmac.Equal(clientHMAC, expectedHMAC) {
		log.Println("Yeah, I've got expectedHMAC!")
	} else {
		log.Println("Oh, No. Wrong HMAC from the client!")
	}

	// parse and save files
	saveFile1, err := getFileFromRequest(r, "file1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	saveFile2, err := getFileFromRequest(r, "file2")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	log.Println("file saved:", saveFile1)
	log.Println("file saved:", saveFile2)

	// 2. response with server side files and data
	// https://peter.bourgon.org/blog/2019/02/12/multipart-http-responses.html
	mw := multipart.NewWriter(w)
	w.Header().Set("Content-Type", mw.FormDataContentType())

	// prepare files and data
	fn1 := "testserver1.txt"
	fn2 := "testserver2.txt"

	randomServerNumber := "9988776655"
	serverHMAC := []byte{0x99,0x99,0x99,0x0e,0xc5,0xd3,0x35,0xd4,0x72,0x4f,0x3e,0xcf,0xfa,0x76,0x6a,0x4c,0x3e,0x07,0x60,0x06,0xa1,0x2b,0x78,0x75,0x52,0x0f,0x80,0x9d,0xde,0x17,0xbf,0xde}
	// 9999990ec5d335d4724f3ecffa766a4c3e076006a12b7875520f809dde17bfde
	hmacBase64 := base64.StdEncoding.EncodeToString(serverHMAC)

	// files
	err = writeFileToMultipartWriter(mw, "file1", fn1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return 
	}
	err = writeFileToMultipartWriter(mw, "file2", fn2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return 
	}
	
	// add key, value
	err = mw.WriteField("data1", randomServerNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return 
	}
	err = mw.WriteField("data2", hmacBase64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return 
	}

	// and close
	mw.Close()
}

func main() {
	http.HandleFunc("/upload", upload)
	// http.Handle("/", http.FileServer(http.Dir("static")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
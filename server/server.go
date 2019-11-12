// https://play.golang.org/p/MrE9BwNbB1

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)


func handler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("received")
	path, _ := os.Getwd()

	// file1
	file1, header1, err := r.FormFile("file1")
	defer file1.Close()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	path1 := path + "/" + header1.Filename

	out, err := os.Create(path1)
	if err != nil {
		fmt.Fprintf(w, "Failed to open the file for writing")
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file1)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	// the header contains useful info, like the original file name
	fmt.Fprintf(w, "File %s uploaded successfully.", header1.Filename)

	// file1
	file2, header2, err := r.FormFile("file2")
	defer file2.Close()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	path2 := path + "/" + header2.Filename

	out2, err := os.Create(path2)
	if err != nil {
		fmt.Fprintf(w, "Failed to open the file for writing")
		return
	}
	defer out2.Close()
	_, err = io.Copy(out2, file2)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	// the header contains useful info, like the original file name
	fmt.Fprintf(w, "File %s uploaded successfully.", header2.Filename)
}


func main() {
	http.ListenAndServe(":8080", http.HandlerFunc(handler))
}
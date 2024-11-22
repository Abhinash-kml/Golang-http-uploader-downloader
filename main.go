package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const uploadDirectory string = "uploaded files"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request IP: ", r.RemoteAddr)
		http.ServeFile(w, r, "./public/index.html")
	})

	http.HandleFunc("/download/{name}", func(w http.ResponseWriter, r *http.Request) {
		fileRequested := r.PathValue("name")
		fmt.Println(fileRequested)

		http.ServeFile(w, r, uploadDirectory+"/"+fileRequested)
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}

		w.Header().Set("Content-Type", "application/x-www-form-urlencoded") // Set the header
		err := r.ParseMultipartForm(10 << 20)                               // Parse the multipart form and set max file limit
		if err != nil {
			LogAndWriteStatusCode(w, http.StatusInternalServerError, "Error parsing form", err)
			return
		}

		multipartFile, multiPartFileHeader, err := r.FormFile("file") // Get the file
		if err != nil {
			LogAndWriteStatusCode(w, http.StatusInternalServerError, "Something went wrong while getting the multipart file", err)
			return
		}
		defer multipartFile.Close()

		fileExtension := strings.ToLower(filepath.Ext(multiPartFileHeader.Filename)) // Get file extension
		filename := multiPartFileHeader.Filename

		//path := filepath.Join(".", "files") // Join filepath
		//fmt.Println(path)
		err = os.MkdirAll(uploadDirectory, os.ModePerm) // Create folders
		if err != nil {
			LogAndWriteStatusCode(w, http.StatusInternalServerError, "Error creating directory for file", err)
			return
		}
		fullpath := uploadDirectory + "/" + filename + fileExtension // Combine to get a valid filepath

		filehandler, err := os.OpenFile(fullpath, os.O_WRONLY|os.O_CREATE, os.ModePerm) // Create file
		if err != nil {
			LogAndWriteStatusCode(w, http.StatusInternalServerError, "Something went wrong while creating file on server", err)
			return
		}
		defer filehandler.Close()

		_, err = io.Copy(filehandler, multipartFile) // Copy contents to the newly created file
		if err != nil {
			LogAndWriteStatusCode(w, http.StatusInternalServerError, "Something went wrong while copying the contents of the file on server", err)
		}
	})

	go func() {
		fmt.Println("----- Starting server on localhost:8000 -----")
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	signal := <-signalChan
	fmt.Println("----- Shutting down server, recieved signal: ", signal)

	// Gracefully shutdown the server here
}

func LogAndWriteStatusCode(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Println(err)
	w.WriteHeader(statusCode)

	if message != "" {
		json.NewEncoder(w).Encode(message)
	}
}

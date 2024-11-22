package main

import (
	"encoding/json"
	"fmt"
	"io"
	"local/models"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	uploadDirectory string = "uploaded files"
	serverPort      string = "8000"
)

// Allowed file extensions, more can be allwed dynamically via POST request
var allowedFileExtensions = []string{
	".jpeg",
	".png",
	".svg",
	".txt",
	".ini",
	".json",
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request IP: ", r.RemoteAddr)
		http.ServeFile(w, r, "./public/index.html")
	})

	http.HandleFunc("/download/{name}", func(w http.ResponseWriter, r *http.Request) {
		fileRequested := r.PathValue("name")
		fmt.Println(fileRequested)

		w.Header().Set("Content-Disposition", "attachment") // Force the requester's browser to download the file
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

		// Check if filetype is allowed
		fileExtension := strings.ToLower(filepath.Ext(multiPartFileHeader.Filename)) // Get file extension
		if !IsExtensionAllowed(fileExtension) {
			http.Error(w, "Uploaded filetype not supported", http.StatusInternalServerError)
			return
		}

		filename := multiPartFileHeader.Filename

		//path := filepath.Join(".", "files") // Join filepath
		//fmt.Println(path)
		err = os.MkdirAll(uploadDirectory, os.ModePerm) // Create folders
		if err != nil {
			LogAndWriteStatusCode(w, http.StatusInternalServerError, "Error creating directory for file", err)
			return
		}
		fullpath := uploadDirectory + "/" + filename /*+ fileExtension*/ // Combine to get a valid filepath
		// Note to self: FileHeader.Filename contains the complete
		// name of the file along with the extension
		// So no need to combine here

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

	http.HandleFunc("/fileextension", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is allowed on this route", http.StatusBadRequest)
			return
		}

		opRequest := &models.OperationRequest{}
		json.NewDecoder(r.Body).Decode(opRequest)

		if opRequest.Operation == "add" {
			allowedFileExtensions = append(allowedFileExtensions, opRequest.Value)
			fmt.Println(allowedFileExtensions)
		} else if opRequest.Operation == "remove" {
			var extensionIndex uint
			for index, value := range allowedFileExtensions {
				if value == opRequest.Value {
					extensionIndex = uint(index)
					break
				}
			}

			fmt.Println("Extensions before removal: ") // Log
			fmt.Println(allowedFileExtensions)         // Log

			newSlice := allowedFileExtensions[:extensionIndex]
			newSlice = append(newSlice, allowedFileExtensions[extensionIndex+1:]...)
			allowedFileExtensions = newSlice

			fmt.Println("Extensions after removal: ") // Log
			fmt.Println(allowedFileExtensions)        // Log
		} else {
			http.Error(w, "Please specify a valid operation between \"add\" or \"remove\". ", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	go func() {
		fmt.Printf("----- Starting server on localhost:%s -----", serverPort)
		log.Fatal(http.ListenAndServe(":"+serverPort, nil))
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	signal := <-signalChan
	fmt.Println("----- Shutting down server, recieved signal: ", signal)

	// Gracefully shutdown the server here
}

func IsExtensionAllowed(extension string) bool {
	for _, val := range allowedFileExtensions {
		if extension == val {
			return true
		}
	}

	return false
}

func LogAndWriteStatusCode(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Println(err)
	w.WriteHeader(statusCode)

	if message != "" {
		json.NewEncoder(w).Encode(message)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	s3 "minioTest"
	"net/http"
	"time"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/upload", upload).Methods(http.MethodPost)
	r.HandleFunc("/download/{objectKey}", download).Methods(http.MethodGet)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// TODO 환경에 따라 환경변수를 셋팅하여 env 값을 넣도록 수정할 것
	result, err := s3.UploadFile(file, fileHeader.Filename, "dev")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectKey := vars["objectKey"]

	// TODO 환경에 따라 환경변수를 셋팅하여 env 값을 넣도록 수정할 것
	fileByte, err := s3.DownloadFile(objectKey, "dev")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, bytes.NewReader(fileByte))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

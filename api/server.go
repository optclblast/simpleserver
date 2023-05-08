package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

const (
	SERVER_RESPONSE         string = "SERVER_RESPONSE"
	SERVER_REQUEST          string = "SERVER_REQUEST"
	CLIENT_GREETING_REQUEST string = "CLIENT_GREETING_REQUEST"
	CLIENT_MEDIA_FILE       string = "CLIENT_MEDIA_FILE"
	CLIENT_STATUS_REQUEST   string = "CLIENT_STATUS_REQUEST"
)

type MediaFile struct {
	Name string
	Link string
	Ext  string
}

type Request struct {
	ID              string `json:"id"`
	RequestCode     string `json:"code"`
	RequestContents string `json:"requestcontents"`
}

type Server struct {
	*mux.Router
	filesList map[int]MediaFile
	requests  []Request
}

func NewServer() *Server {
	server := &Server{
		Router:    mux.NewRouter(),
		filesList: map[int]MediaFile{},
		requests:  []Request{},
	}
	server.routes()
	return server

}

func (server *Server) routes() {
	go server.HandleFunc("/server-status", server.serverStatus()).Methods("GET")
	go server.HandleFunc("/getfile", server.receiveFileFromClient()).Methods("POST")
}

func (server *Server) serverStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var json_content Request
		err := json.NewDecoder(r.Body).Decode(&json_content)
		if err != nil || json_content.ID == "" || json_content.RequestCode != CLIENT_GREETING_REQUEST {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json_content.ID = "serv-0001"
		json_content.RequestCode = SERVER_RESPONSE
		json_content.RequestContents = "The server is OK"
		server.requests = append(server.requests, json_content)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(json_content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (server *Server) receiveFileFromClient() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		err := request.ParseMultipartForm(64 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		file, h, err := request.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tmpfile, err := os.Create("./files/" + fmt.Sprintf("%s ", h.Filename)) //saving file to files dir
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("An error has accured during creating file\n%s \n", err)
			return
		}

		n, err := io.Copy(tmpfile, file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("An error has accured during copying data to created file\n%s \n", err)
			return
		}
		fmt.Printf("Recieved new file. \n Its content:\n%d\n", n)
		defer tmpfile.Close()

		//Adding recieved file in the queue
		file_in_queue := len(server.filesList)
		server.filesList[file_in_queue] = MediaFile{
			Name: h.Filename,
			Link: fmt.Sprintf("./files/%s", h.Filename),
			Ext:  strings.Trim(filepath.Ext(h.Filename), "."),
		}

		w.WriteHeader(204)
	}
}

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gorilla/mux"
)

const (
	whisperAIScript = "./script.py" //whisper-ai script
	downloadingDir  = "./files"     //filder to store our files
	//file statuses
	ACCEPTED    = "ACCEPTED"
	IN_PROGRESS = "IN_PROGRESS"
	DONE        = "DONE"
)

// struct to store data about file and request
type MediaFile struct {
	GUID   string
	Name   string
	Path   string
	Status string
	Link   string
	WH     string
}

// request itself
type Request struct {
	ID            string `json:"id"`
	RequestStatus string `json:"status"`
}

// server itself
type Server struct {
	*mux.Router
	filesList map[string]MediaFile
}

// server init
func NewServer() *Server {
	server := &Server{
		Router:    mux.NewRouter(),
		filesList: map[string]MediaFile{},
	}
	server.routes()
	return server
}

// servers worldmap
func (server *Server) routes() {
	go server.HandleFunc("/server-status", server.serverStatus()).Methods("GET")
	go server.HandleFunc("/getAll", server.getAll()).Methods("GET")
	go server.HandleFunc("/getFile", server.getFile()).Methods("POST")
	go server.HandleFunc("/postFile", server.receiveFileFromClient()).Methods("POST")
	//FK server endpoints
	go server.HandleFunc("/send-link", server.handleSendLink()).Methods("GET")
	go server.HandleFunc("/get-result", server.handleGetResult()).Methods("GET")

}

// endpoint handler
func (server *Server) handleSendLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//reading from query string
		params := r.URL.Query()
		fileID := params.Get("file_id")
		link := params.Get("link")
		responseURL := params.Get("response_uri")

		fmt.Printf("PARAMS: %s\n", params)
		//handling missing query string
		if fileID == "" || responseURL == "" || link == "" {

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//Downloading file via link in the gotten request
		resp, err := http.Get(link)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("%v\n", resp)

		fileName := filepath.Base(resp.Request.URL.Path)
		newdir := fmt.Sprintf("%s\\%s", downloadingDir, fileName)
		err = os.Mkdir(newdir, 0750)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("An error has accured during creating a folder\n%s \n", err)
			return
		}
		if fileName[len(fileName)-4:] != ".mp3" {
			fileName += ".mp3"
		}
		file, err := os.Create(fmt.Sprintf("%s\\", newdir) + fileName) //saving file to files dir
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("An error has accured during creating file\n%s \n", err)
			return
		}

		//Writing data to a file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("An error has accured during writing a file\n%s \n", err)
			return
		}

		file_inlist := MediaFile{
			GUID:   fileID,
			Path:   newdir,
			Name:   fileName,
			Status: ACCEPTED,
			Link:   link,
			WH:     responseURL,
		}
		log.Printf("FILE: %s\n", file_inlist) //logging recieved file

		server.filesList[fileID] = file_inlist //adding file into a table of files, that server cares of at the moment

		defer resp.Body.Close()

		//executing whisper's scpipt
		go evaluatingWhisper(fmt.Sprintf("%s\\%s", newdir, fileName), server.filesList[fileID])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("An error has accured during evaling whisper\n%s \n", err)
			return
		}
		w.WriteHeader(200)
	}
}

func (server *Server) handleGetResult() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		fileID := params.Get("fileID")

		//checking for file in the list
		if _, ok := server.filesList[fileID]; !ok {
			w.WriteHeader(http.StatusNotFound) //not found
			return
		}

		//got the data
		result := server.filesList[fileID]
		text, tone, summary := text_tone_summary(result)

		//prepearing response with out content
		response := map[string]string{
			"text":    string(text),
			"tone":    string(tone),
			"summary": string(summary),
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the response JSON to the response writer
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(responseJSON)
	}
}

func (server *Server) serverStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var json_content Request
		err := json.NewDecoder(r.Body).Decode(&json_content)
		if err != nil || json_content.ID == "" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json_content.ID = "correct"
		//json_content.RequestContents = "The server is OK"
		//server.requests = append(server.requests, json_content)

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
			log.Printf("An error has accured during creating file\n%s \n", err)
			return
		}

		n, err := io.Copy(tmpfile, file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("An error has accured during copying data to created file\n%s \n", err)
			return
		}
		log.Printf("Recieved new file. \n Its content:\n%d\n", n)
		defer tmpfile.Close()

		w.WriteHeader(204)
	}
}

func (server *Server) getAll() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		var json_content Request
		err := json.NewDecoder(request.Body).Decode(&json_content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		//todo!()
	}
}

func (server *Server) getFile() http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		//todo!()
	}
}

func evaluatingWhisper(file_path string, fileinfo MediaFile) error {
	//Need to do flag eval
	log.Printf("PATH: %s\n", file_path)
	fileinfo.Status = IN_PROGRESS
	cmd := exec.Command("python", whisperAIScript, file_path)
	err := cmd.Run()
	if err != nil {
		return err
	}
	for {
		_, err_txt := os.Stat(fmt.Sprintf("%s\\%s_text.txt", fileinfo.Path, fileinfo.Name))
		//_, err_tone := os.Stat(fmt.Sprintf("%s\\%s_tone", fileinfo.Path, fileinfo.Name))
		//_, err_summary := os.Stat(fmt.Sprintf("%s\\%s_summary", fileinfo.Path, fileinfo.Name))
		if err_txt != nil { //&& err_tone != nil && err_summary != nil {
			break
		}

	}
	fileinfo.Status = DONE
	pingClientFK(fileinfo)
	return nil
}

func pingClientFK(file MediaFile) ([]byte, error) {
	var json_content Request
	json_content.ID = file.GUID
	json_content.RequestStatus = file.Status

	jsonValue, _ := json.Marshal(json_content)

	req, err := http.NewRequest(http.MethodPost, file.WH, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Fatal("wtf json")
	}

	client := &http.Client{}
	response := &http.Response{}
	for {
		res, err := client.Do(req)
		if res.StatusCode != http.StatusNotFound {
			response = res
			break
		}

		if err != nil {
			return []byte{}, err
		}
	}
	defer response.Body.Close()
	log.Println(response.StatusCode)
	cnt, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}
	return cnt, nil
}

// reading from txt files
func text_tone_summary(file MediaFile) ([]byte, []byte, []byte) {
	text := readFile(fmt.Sprintf("%s\\%s_text", file.Path, file.Name))
	tone := readFile(fmt.Sprintf("%s\\%s_tone", file.Path, file.Name))
	summary := readFile(fmt.Sprintf("%s\\%s_summary", file.Path, file.Name))
	return text, tone, summary
}

func readFile(filetoread string) []byte {
	file_info, error := os.Stat(filetoread)
	if error != nil {
		log.Fatal(error)
	}
	file_size := file_info.Size()
	file, err := os.Open(filetoread)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	buf := make([]byte, file_size)
	file.Read(buf)
	return buf
}

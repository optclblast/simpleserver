package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"server/db"
	"time"

	"github.com/gorilla/mux"
)

var API_KEY_CLIENT = []string{"11", "22", "223", "444"}
var API_KEY_VIP = []string{"555"}

const (
	OPENAI_API_TOKEN = "TOKEN"

	whisperAIScript = "./script.py" //whisper-ai script
	downloadingDir  = "./files"
	DONE            = "DONE"
	ACCEPTED        = "ACCEPTED"
	IN_PROGRESS     = "IN_PROGRESS"
)

type MediaFile struct {
	GUID        string
	Name        string
	Path        string
	Status      string
	Link        string
	WH          string
	WhisperDone bool
}

type Request struct {
	ID            string `json:"id"`
	RequestStatus string `json:"status"`
}

type LogEntry struct {
	date     time.Time
	contents string
}

type Server struct {
	*mux.Router
	filesList map[string]MediaFile
	bussy     bool
}

func NewLogEntry(date time.Time, content string) LogEntry {
	logentry := LogEntry{
		date:     date,
		contents: content,
	}
	return logentry
}

func NewServer() *Server {
	server := &Server{
		Router:    mux.NewRouter(),
		filesList: make(map[string]MediaFile),
		bussy:     false,
	}
	fmt.Println(server.bussy)
	server.routes()
	return server
}

func (server *Server) routes() {
	server.HandleFunc("/send-link", server.handleSendLink()).Methods("GET")
	server.HandleFunc("/get-result", server.handleGetResult()).Methods("GET")
	server.HandleFunc("/whisper_ping", server.handleWhisperPing()).Methods("GET")
	server.HandleFunc("/signin", server.handleMainPage()).Methods("GET")
	server.HandleFunc("/userPage", server.getUserPage()).Methods("GET")
	server.HandleFunc("/userPage/data/getfile", server.fileReaderPage()).Methods("GET")
	/////////////////
	server.HandleFunc("/userPage/data", server.getUserData()).Methods("POST")
	server.HandleFunc("/userPage/data/getfile", server.getFileReq()).Methods("POST")
	server.HandleFunc("/main/diarize-file", server.handleUploadedFile()).Methods("POST")
	server.HandleFunc("/signin/auth", server.handleSignIn()).Methods("POST")
	server.HandleFunc("/register", server.handleRegister()).Methods("POST")
	/////////////////
	server.HandleFunc("/userPage/data/deletefile", server.deleteUserFile()).Methods("DELETE")
}

// recieving data from client
func (server *Server) handleSendLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		fileID := params.Get("file_id")
		apiKey := params.Get("api_key")
		link := params.Get("link")
		responseURL := params.Get("response_uri")
		fmt.Println(server.bussy)

		//auth
		if !ValidKey(API_KEY_CLIENT, apiKey) {
			w.WriteHeader(http.StatusUnauthorized)
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Invalid authorization at /send-link\nApi key: %s\nRequest by: %s", apiKey, responseURL),
			})
			return
		}

		//handling missing query string
		if fileID == "" || responseURL == "" || link == "" {
			w.WriteHeader(http.StatusBadRequest)
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Invalid request at /send-link by: %s", responseURL),
			})
			return
		}

		//Downloading file via link in the gotten request
		resp, err := http.Get(link)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Cannot download file by link: %s", link),
			})
			return
		}
		fmt.Println(resp)
		fileName := removeSpaces(filepath.Base(resp.Request.URL.Path))
		newdir := fmt.Sprintf("%s/%s", downloadingDir, fileName)
		err = os.Mkdir(newdir, 0750)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("An error has accured during creating a folder (/send-link)\n%s \n", err),
			})
			return
		}

		if fileName[len(fileName)-4:] != ".mp3" {
			fileName += ".mp3"
		}

		file, err := os.Create(fmt.Sprintf("%s/", newdir) + fileName) //saving file to files dir
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("An error has accured during creating file\n%s (/send-link)\n", err),
			})
			return
		}

		//Writing data to a file
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("An error has accured during writing a file\n%s \n", err),
			})
			return
		}

		file_inlist := MediaFile{
			GUID:        fileID,
			Path:        newdir,
			Name:        fileName,
			Status:      ACCEPTED,
			Link:        link,
			WH:          responseURL,
			WhisperDone: false,
		}
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprint("File has been created: \n", file_inlist),
		})
		server.filesList[fileID] = file_inlist
		defer resp.Body.Close()
		defer file.Close()

		//executing whisper's scpipt
		go server.evaluatingWhisper(fmt.Sprintf("%s/%s", newdir, fileName), server.filesList[fileID])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("An error has accured during evaling whisper\n%s \n", err),
			})
			return
		}
		w.WriteHeader(200)
	}
}

// sending result back to client
func (server *Server) handleGetResult() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		fileID := params.Get("file_id")
		if apiKey := params.Get("apiKey"); ValidKey(API_KEY_CLIENT, apiKey) {
			w.WriteHeader(http.StatusUnauthorized) //
			return
		}

		file := server.filesList[fileID]

		text, tone, summary := textToneSummary(file)

		response := map[string]string{
			"text":    string(text),
			"tone":    string(tone),
			"summary": string(summary),
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Error at json marshalling (handleGetResult):\n%s \n", err),
			})
			return
		}

		// Write the response JSON to the response writer
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(responseJSON)

		err = os.RemoveAll(file.Path)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Error at files removing (handleGetResult):\n%s \n", err),
			})
		}
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprintf("A file (%s) sended to %s successfully!", file.GUID, file.WH),
		})
		delete(server.filesList, file.GUID)
	}
}

// recieving ping from whisper that file (by guid) is done or script ended up with errors
func (server *Server) handleWhisperPing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		fileID := params.Get("file_id")
		errors := params.Get("error")

		file := server.filesList[fileID]

		dbconn, err := db.NewConnection()
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		defer dbconn.Close()
		qry := db.New(dbconn)
		dbfile, err := qry.GetFileByGuid(context.Background(), fileID)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprint(err),
			})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors != "" {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Whisper cannot handle %s request! Whisper's error:\n%s\n", fileID, errors),
			})
			w.WriteHeader(http.StatusBadRequest)
			server.pingCLErrorCase(file, errors)
			return
		}

		if file.Link == "web" {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("File from client: %s", file.Name),
			})
		}

		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprintf("Whisper finished working on %s file\n", fileID),
		})

		for key := range server.filesList {
			fmt.Println(key)
		}
		file.WhisperDone = true

		updateFileParams := db.UpdateFileParams{
			Id:     dbfile.Id,
			Name:   dbfile.Name,
			Status: "DONE",
		}
		_, err = qry.UpdateFile(context.Background(), updateFileParams)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Cannot update file %s", err),
			})
		}

		file.Status = DONE
		server.bussy = false
		b, err := pingClientVIP(file)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Bad req to cl! %s\n%d", err, b),
			})
		}
	}
}

// evoke whisper script
func (server *Server) evaluatingWhisper(file_path string, fileinfo MediaFile) error {
	//Need to do flag eval
	fmt.Println(server.bussy)
	fmt.Printf("PATH: %s\n", file_path)
	fileinfo.Status = "transcribing"
	for {
		if !server.bussy {
			server.bussy = true
			break
		}

		time.Sleep(time.Minute)
		fmt.Printf("Waiting for server's queue resolving... Current state: %v\n", server.bussy)
	}
	fmt.Println("Summoning whisper")
	cmd := exec.Command("python", whisperAIScript, file_path, fileinfo.GUID)
	err := cmd.Run()

	if err != nil {
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprintf("Error at running Whisper:\n%s \n", err),
		})
		return err
	}
	return nil
}

// notificate client about file complition
func pingClientVIP(file MediaFile) ([]byte, error) {
	var json_content Request
	json_content.ID = file.GUID
	json_content.RequestStatus = file.Status

	jsonValue, _ := json.Marshal(json_content)

	req, err := http.NewRequest(http.MethodPost, file.WH, bytes.NewBuffer(jsonValue))
	if err != nil {
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprintf("Error at sending ping request to %s, got the following error:\n%s", file.WH, err),
		})
	}

	client := &http.Client{}
	response := &http.Response{}
	for {
		res, err := client.Do(req)
		if res.StatusCode != http.StatusNotFound {
			response = res
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Successfully sanded data to %s!", file.WH),
			})
			break
		}

		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Error at sending ping request to %s (while re-trying), got the following error:\n%s", file.WH, err),
			})
			return []byte{}, err
		}
	}
	defer response.Body.Close()
	fmt.Println(response.StatusCode)
	cnt, err := io.ReadAll(response.Body)
	if err != nil {
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprintf("Error at response from %s, got the following error:\n%s", file.WH, err),
		})
		return []byte{}, err
	}
	return cnt, nil
}

func (server *Server) pingCLErrorCase(file MediaFile, errmsg string) ([]byte, error) {
	var json_content Request
	json_content.ID = file.GUID
	json_content.RequestStatus = "При обработке записи разговора произошла ошибка."
	jsonValue, _ := json.Marshal(json_content)

	req, err := http.NewRequest(http.MethodPost, file.WH, bytes.NewBuffer(jsonValue))
	if err != nil {
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprint("Error at sending request at:", file.WH, "\n", "Error:", "\n", err, "\n"),
		})
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
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprint("Error at sending request at:", file.WH, "\n", "Error:", "\n", err, "\n"),
			})
			return []byte{}, err
		}
	}
	defer response.Body.Close()
	fmt.Println(response.StatusCode)
	cnt, err := io.ReadAll(response.Body)
	if err != nil {
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprint("Error at reading response:\n", err, "\n"),
		})
		return []byte{}, err
	}

	log.Println(errmsg)
	err = os.RemoveAll(file.Path)
	if err != nil {
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprint("Error at removing files:\n", err, "\n"),
		})
	}
	delete(server.filesList, file.GUID)
	return cnt, nil
}

// collecting txt files data
func textToneSummary(file MediaFile) ([]byte, []byte, []byte) {
	fmt.Printf("%s/%s_text.txt\n", file.Path, file.Name)
	text := readFile(fmt.Sprintf("%s/%s_text.txt", file.Path, file.Name))
	tone := text
	summary := text
	return text, tone, summary
}

// reading data from file
func readFile(filetoread string) []byte {
	fmt.Printf("%s\n", filetoread)
	file_info, error := os.Stat(filetoread)
	if error != nil {
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprint("Error at os.Stat:\n", error, "\n"),
		})
	}
	file_size := file_info.Size()
	file, err := os.Open(filetoread)
	if err != nil {
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprintf("Error at reading from file (%s):\n%s", filetoread, err),
		})
	}
	defer func() {
		if err = file.Close(); err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Error at closing file (%s):\n%s", filetoread, err),
			})
		}
	}()
	buf := make([]byte, file_size)
	file.Read(buf)
	return buf
}

// checking if the key is valid
func ValidKey(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// writing a log entry
func Logger(entry LogEntry) {
	date := fmt.Sprintf("%d.%d.%d", entry.date.Day(), entry.date.Month(), entry.date.Year())
	filename := fmt.Sprintf("logby_%s.log", date)
	f, err := os.OpenFile("./logs/"+filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)
	log.Print("				" + entry.contents)
}

func removeSpaces(name string) string {
	runes := []rune(name)
	for i, c := range runes {
		if c == ' ' {
			runes[i] = '_'
		}
	}
	return string(runes)
}

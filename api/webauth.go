package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"server/auth"
	"server/db"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
)

type FileRow struct {
	Id     int64
	Name   string
	Status string
	Path   string
}

type ListJsonPattern struct {
	Files []db.File `json:"files"`
}

var store = sessions.NewCookieStore([]byte(os.Getenv("SERVER_SESSION_KEY")))

func (server *Server) handleMainPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		userSession, ok := session.Values["login"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dbconn, err := db.NewConnection()
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		defer dbconn.Close()
		qry := db.New(dbconn)

		_, err = qry.GetAccountBySession(context.Background(), userSession)
		if err != nil {
			w.Write(readFile("./htmls/loginregister.html"))
			return
		}

		http.Redirect(w, r, "http://server/userPage", http.StatusAccepted)
	}
}

func (server *Server) handleSignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerContentTtype := r.Header.Get("Content-Type")
		if headerContentTtype != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		r.ParseForm()

		var login, password string
		for key, value := range r.Form {
			if key == "login" {
				login = value[0]
			} else if key == "password" {
				password = value[0]
			}
		}

		logUser := auth.HashAuthData(login, password)

		dbconn, err := db.NewConnection()
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}

		qry := db.New(dbconn)
		defer dbconn.Close()

		sysUser, err := qry.GetAccountByLogin(context.Background(), logUser.Login)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "Wrong login or password", http.StatusBadRequest)
			return
		}

		if logUser.Password != sysUser.Password {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "Wrong login or password", http.StatusBadRequest)
			return
		}

		session, err := store.Get(r, "session")
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
		}

		session.Values["login"] = login

		updareAccQuery := db.UpdateAccountParams{
			Id:       sysUser.Id,
			Login:    sysUser.Login,
			Password: sysUser.Password,
			Address:  sysUser.Address,
			Session:  login,
		}

		_, err = qry.UpdateAccount(context.Background(), updareAccQuery)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = session.Save(r, w)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//fmt.Printf("User %s signd in!\n", hex.EncodeToString(user.Login[:]))
		w.WriteHeader(http.StatusOK)
	}
}

func (server *Server) handleRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headerContentTtype := r.Header.Get("Content-Type")
		if headerContentTtype != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		r.ParseForm()
		var login, password string
		for key, value := range r.Form {
			if key == "login" {
				login = value[0]
			} else if key == "password" {
				password = value[0]
			}
		}
		fmt.Printf("%s\n%s\n", login, password)
		user := auth.HashAuthData(login, password)

		dbconn, err := db.NewConnection()
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}

		qry := db.New(dbconn)
		defer dbconn.Close()

		_, err = qry.GetAccountByLogin(context.Background(), user.Login)
		if err == nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprintf("%s, loc: [reg]GetAccountByLogin()", err)))
			http.Error(w, "account already exists", http.StatusInternalServerError)
			return
		}

		lastid, err := qry.GetLastAccountID(context.Background())
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprintf("%s, loc: [reg]GetLastAccountID()", err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		arg := db.CreateAccountParams{
			Id:        lastid + 1,
			Login:     user.Login,
			Password:  user.Password,
			Address:   "web",
			CreatedAt: time.Now(),
			Session:   "",
		}
		_, err = qry.CreateAccount(context.Background(), arg)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprintf("%s, loc: [reg]CreateAccount()", err)))
			http.Error(w, "cannnot create account", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (server *Server) getUserPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		_, ok := session.Values["login"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err := w.Write(readFile("./htmls/lk.html"))
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			fmt.Printf("herres the error!\n%s", err)
		}
	}
}

func (server *Server) getUserData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		userSession, ok := session.Values["login"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		dbconn, err := db.NewConnection()
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		defer dbconn.Close()
		qry := db.New(dbconn)

		user, err := qry.GetAccountBySession(context.Background(), userSession)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprint(err),
			})
		}

		listFilesParam := db.ListFilesParams{
			Limit:  100,
			Offset: 0,
		}

		files, err := qry.ListFiles(context.Background(), listFilesParam, user.Id)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprint(err),
			})
		}

		body := ListJsonPattern{Files: files}
		json, err := json.Marshal(body)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprint(err),
			})
		}
		w.Write(json)
	}
}

func (server *Server) getFileReq() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (server *Server) fileReaderPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileid, err := strconv.Atoi(r.URL.Query().Get("fid"))
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}

		session, _ := store.Get(r, "session")
		userSession, ok := session.Values["login"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			Logger(NewLogEntry(time.Now(), "There are no such field `login` in the given session"))
			return
		}
		dbconn, err := db.NewConnection()
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		defer dbconn.Close()
		qry := db.New(dbconn)

		user, err := qry.GetAccountBySession(context.Background(), userSession)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}

		file, err := qry.GetFile(context.Background(), int64(fileid), user.Id)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "404", http.StatusNotFound)
			return
		}

		w.Write(readFile(file.LocationTxt))
	}
}

func (server *Server) handleUploadedFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(server.bussy)
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		session, _ := store.Get(r, "session")
		userSession, ok := session.Values["login"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			Logger(NewLogEntry(time.Now(), fmt.Sprint("There are no such field `login` in the given session")))
			return
		}
		dbconn, err := db.NewConnection()
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		defer dbconn.Close()
		qry := db.New(dbconn)

		user, err := qry.GetAccountBySession(context.Background(), userSession)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprint(err),
			})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprint(err),
			})
			http.Error(w, "Failed to retrieve file", http.StatusBadRequest)
			return
		}
		defer file.Close()
		name := removeSpaces(handler.Filename)
		dst, err := os.Create(".\\services\\httpserver_fk\\fk_files\\" + name)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprint(err),
			})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprint(err),
			})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		Logger(LogEntry{
			date:     time.Now(),
			contents: "File uploaded successfully",
		})

		file_inlist := MediaFile{
			GUID:        "fk" + name,
			Path:        ".\\services\\httpserver_fk\\fk_files",
			Name:        name,
			Status:      ACCEPTED,
			Link:        "fk",
			WH:          "fk",
			WhisperDone: false,
			SummaryDone: false,
			ToneDone:    false,
		}
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprint("File has been created: \n", file_inlist),
		})

		lastid, err := qry.GetLastFileID(context.Background())
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprintf("%s, loc: [handleUploadedFile]GetLastFileID()", err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		createFileParams := db.CreateFileParams{
			Id:          lastid + 1,
			Owner:       user.Id,
			Name:        file_inlist.Name,
			Location:    file_inlist.Path + "\\" + file_inlist.Name,
			LocationWav: file_inlist.Path + "\\" + strings.Split(filepath.Base(file_inlist.Name), ".")[0] + ".wav",
			LocationTxt: file_inlist.Path + "\\" + strings.Split(filepath.Base(file_inlist.Name), ".")[0] + ".mp3_text.txt",
			CreatedAt:   time.Now(),
			Status:      file_inlist.Status,
			Guid:        file_inlist.GUID,
		}

		_, err = qry.CreateFile(context.Background(), createFileParams)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprintf("%s, loc: [handleUploadedFile]CreateFile()", err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}

		server.filesList["fk"+handler.Filename] = file_inlist
		go server.evaluatingWhisper(fmt.Sprintf("%s\\%s", ".\\httpserver_fk\\fk_files", name), server.filesList["fk"+name])
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

func (server *Server) deleteUserFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileid, err := strconv.Atoi(r.URL.Query().Get("fid"))
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}

		session, _ := store.Get(r, "session")
		userSession, ok := session.Values["login"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			Logger(NewLogEntry(time.Now(), "There are no such field `login` in the given session"))
			return
		}
		dbconn, err := db.NewConnection()
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		defer dbconn.Close()
		qry := db.New(dbconn)

		user, err := qry.GetAccountBySession(context.Background(), userSession)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}

		file, err := qry.GetFile(context.Background(), int64(fileid), user.Id)
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}

		err = os.RemoveAll(file.Location)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Error at files removing (deleteUserFile):\n%s \n", err),
			})
		}
		err = os.RemoveAll(file.LocationWav)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Error at files removing (deleteUserFile):\n%s \n", err),
			})
		}
		err = os.RemoveAll(file.LocationTxt)
		if err != nil {
			Logger(LogEntry{
				date:     time.Now(),
				contents: fmt.Sprintf("Error at files removing (deleteUserFile):\n%s \n", err),
			})
		}
		Logger(LogEntry{
			date:     time.Now(),
			contents: fmt.Sprintf("A files %s || %s || %s have been deleted successfully!", file.Location, file.LocationWav, file.LocationTxt),
		})

		delete(server.filesList, file.Guid)

		err = qry.DeleteFile(context.Background(), int64(fileid))
		if err != nil {
			Logger(NewLogEntry(time.Now(), fmt.Sprint(err)))
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
	}
}

package utils

import (
	"fmt"
	"os"
)

// mock at the moment
type FileRow struct {
	Id     int64
	Name   string
	Status string
	Path   string
}

func GenerateFileListPage(user string) {
	f, err := os.Create(fmt.Sprintf("./htmls/TMPPAGE%s.html", user))
	if err != nil {
		fmt.Printf("Error at creating html: %s\n", err)
		return
	}
	filesTable := []FileRow{
		{1, "file1.txt", "DONE", "./files"},
		{2, "file2.txt", "Error", "./files"},
		{3, "file3.txt", "In progress", "./files"},
		{4, "file4.txt", "DONE", "./files"},
		{5, "file5.txt", "Error", "./files"},
	}

	dataChunks := []string{`<!DOCTYPE html><html lang="ru"><head><meta charset="UTF-8"><title>File uploader</title></head><body><label>Ваши файлы</label><br><br>`, ``, `<script>`, ``, `</script></body></html>`}
	for _, row := range filesTable {
		dataChunks[1] = dataChunks[1] + "<br>" + fmt.Sprintf("%v  %s  %s  ", row.Id, row.Name, row.Status)
		if row.Status == "DONE" {
			dataChunks[1] = dataChunks[1] + fmt.Sprintf("<button id = 'id' type='button' onclick='download%v()'>Посмотреть</button>", row.Id)
			dataChunks[3] = dataChunks[3] + fmt.Sprintf("function download%v() {fetch('http://81.200.28.240:27100/fileview?' + new URLSearchParams({fileID: %v}))}", row.Id, row.Id)
		}
	}
	var data string
	for _, s := range dataChunks {
		data = data + s
	}

	_, err = f.Write([]byte(data))
	if err != nil {
		fmt.Printf("Error at writing html: %s\n", err)
		return
	}
	f.Close()
}

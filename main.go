package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"server/api"
	"time"
)

func main() {
	port := portDeclaration()
	server := api.NewServer()
	addr := "0.0.0.0"
	fmt.Println(addr + port)
	err := http.ListenAndServe(addr+port, server)
	if err != nil {
		api.Logger(api.NewLogEntry(
			time.Now(),
			fmt.Sprintf("Server cannot serve to choosen port (%s)! \n Got the following error:\n%s", port, err),
		))
	}
}

// Выбор порта для сервера
func portDeclaration() string {
	fmt.Println("Please, enter a port the server will be listening to (type \"d\" for 8080 )")

	reader := bufio.NewReader(os.Stdin)
	port, _ := reader.ReadString('\n')

	port = port[:len(port)-2]
	if port == "d" {
		port = "8080"
	}

	api.Logger(api.NewLogEntry(
		time.Now(),
		fmt.Sprintf("Serving to :%s\n", port),
	))

	return fmt.Sprintf(":%s", port)
}

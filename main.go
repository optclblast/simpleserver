package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/api"
)

// 192.168.2.110
func main() {
	port := portDeclaration()
	server := api.NewServer()
	addr := "192.168.2.110"
	fmt.Println(addr + port)
	err := http.ListenAndServe(addr+port, server)
	if err != nil {
		log.Panicf("Server cannot serve to choosen port (%s)! \n Got the following error:\n%s", port, err)
	}
}

// Выбор порта для сервера
func portDeclaration() string {
	fmt.Println("Please, enter a port the server will be listening to")
	reader := bufio.NewReader(os.Stdin)
	port, _ := reader.ReadString('\n')
	port = port[:len(port)-2]
	fmt.Printf("Serving to :%s\n", port)
	return fmt.Sprintf(":%s", port)
}

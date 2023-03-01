/**
 * 20184754 kim-hyunju
 * P2POmokServer.go
 **/

package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var wait bool = false
var client_conn [2]net.Conn
var client_list [2]string
var client_nickname_arr [2]string
var exit int

func main() {

	serverPort := "54754"
	count := 0
	exit = 0

	listener, _ := net.Listen("tcp", ":"+serverPort) // server started
	defer listener.Close()                           //if main terminated, listener close
	SetupCloseHandler()                              // ctrl+c handler

	for {
		conn, _ := listener.Accept()
		count += 1
		go handleConnection(conn, count)
	}
}

func handleConnection(conn net.Conn, count int) {
	SetupCloseHandler()
	defer conn.Close()

	//get client info
	client_address := conn.RemoteAddr().String()
	client_ip := strings.Split(client_address, ":")[0]
	client_TCP_port := strings.Split(client_address, ":")[1]

	for {
		// get client msg
		buffer := make([]byte, 1024)
		//count, _ := conn.Read(buffer)
		conn.Read(buffer)
		now_count := (count - exit) % 2
		opp_count := (count - exit + 1) % 2
		// parsing msg
		if strings.HasPrefix(string(buffer), "exit") {
			fmt.Println(client_nickname_arr[now_count] + " exited.")
			fmt.Println()
			_conn := client_conn[now_count]
			_conn.Close()
			wait = false
			exit += 1
			break
		}
		if strings.HasPrefix(string(buffer), ",") {
			slice := strings.Split(string(buffer), ",")
			client_nickname := slice[1]
			client_nickname_arr[(now_count)] = client_nickname
			client_conn[(now_count)] = conn
			client_UDP_port := slice[2]

			// waiting or not
			if wait == false {
				wait = true
				//client_conn[0] = conn
				client_list[now_count] = "," + client_nickname + "," + client_UDP_port + "," + client_ip + ",1"

				response := "!wait|"
				conn.Write([]byte(response))
				msg := client_nickname + " joined from " + client_ip + ":" + client_TCP_port + ". UDP port " + client_UDP_port + ". \n" + "1 user connected, waiting for another \n"
				fmt.Println(msg)
			} else {
				wait = false
				//client_conn[1] = conn
				// now = 1
				client_list[now_count] = "," + client_nickname + "," + client_UDP_port + "," + client_ip + ",2"

				response := client_list[opp_count] + "|"
				conn.Write([]byte(response))

				response = client_list[now_count] + "|"
				client_conn[opp_count].Write([]byte(response))

				opp_list := strings.Split(client_list[opp_count], ",")
				opp_name := opp_list[opp_count]

				msg := client_nickname + " joined from " + client_ip + ":" + client_TCP_port + ". UDP port " + client_UDP_port + ". \n" + "2 user connected, notifying " + opp_name + " and " + client_nickname + ".\n" + opp_name + " and " + client_nickname + " disconnected.\n\n"
				fmt.Println(msg)

				disconnect_clients(client_conn)
				break
			}
		}

	}
}
func disconnect_clients(client_conn [2]net.Conn) {
	client_conn[0].Close()
	client_conn[1].Close()
}

// ctrl+c handler
func SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("Bye bye~ \n")
		os.Exit(0)
	}()
}

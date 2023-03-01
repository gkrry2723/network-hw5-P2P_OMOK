/**
 * 20184754 kim-hyunju
 * P2POmokClient.go
 **/

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	SERVER_IP = "127.0.0.1"
	//SERVER_IP   = "165.194.35.202"
	SERVER_PORT = "54754"
)

const (
	Row = 10
	Col = 10
)

type Board [][]int

var x, y, turn_count, win, turn int
var my_turn, end_game bool
var ch chan int = make(chan int)
var board = Board{}
var client_name string

//var opp_address net.Addr
var which_conn string

//var conn net.Conn

//var pconn net.PacketConn

func main() {
	// get client name
	if len(os.Args) < 2 {
		fmt.Println("Please enter arg - your nickname!")
		os.Exit(1)
	}
	client_name := os.Args[1]
	which_conn = "conn"

	// ctrl+c handler
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// UDP connection
	pconn, _ := net.ListenPacket("udp", ":")
	localAddr := pconn.LocalAddr().(*net.UDPAddr)
	client_udp_port := localAddr.Port

	// TCP connection
	conn, err := net.Dial("tcp", SERVER_IP+":"+SERVER_PORT)
	if err != nil {
		// error handling : when the server not started yet
		//fmt.Println(err)
		//if err.Error() == "dial tcp 127.0.0.1:54754: connect: connection refused" {
		if err.Error() == "dial tcp 165.194.35.202:54754: connect: connection refused" {
			fmt.Println()
			fmt.Println("The server is not connected.\n")
			fmt.Printf("Start server first.\n")
			os.Exit(0)
		} else { // other unexperted error handling
			fmt.Println("An unexpected error occurred.\n")
			fmt.Printf("Terminate client.\n")
			os.Exit(0)
		}
	}
	SetupCloseHandler1(conn)
	opp_name, opp_UDP_port, opp_IP, _turn := firstTCP(client_name, client_udp_port, pconn, conn)
	turn = _turn
	//fmt.Println(opp_name+opp_UDP_port+opp_IP, turn)
	//fmt.Println(localAddr)

	// play Omok
	for i := 0; i < Row; i++ {
		var tempRow []int
		for j := 0; j < Col; j++ {
			tempRow = append(tempRow, 0)
		}
		board = append(board, tempRow)
	}
	fmt.Println()
	printBoard(board)
	end_game = false
	opp_address, _ := net.ResolveUDPAddr("udp", opp_IP+":"+opp_UDP_port)
	SetupCloseHandler2(pconn, opp_address)
	fmt.Println(opp_address)

	go handleRecMsg(pconn, opp_name, opp_address)

	for {
		terminate := handleSendMsg(pconn, opp_address)
		if terminate == 1 {
			fmt.Printf("Bye~ \n")
			pconn.Close() // disconnection
			os.Exit(0)    // finish client app
			break
		}
	}

	// handling error
	select {
	case <-ch:
		fmt.Println("[Notice] server error, please reconnect")
		os.Exit(2)
	default:
	}

	// ctrl + c handler
	<-sigCh
	fmt.Println("Bye~\n")
	pconn.WriteTo([]byte("3|"), opp_address)
	pconn.Close()
	os.Exit(0)

}

func handleSendMsg(pconn net.PacketConn, opp_address net.Addr) int {
	// 오목 관련 보내기, 채팅
	for {
		var input string

		// for special command
		input_, _ := bufio.NewReader(os.Stdin).ReadString('\n') // get input
		input = strings.TrimSpace(input_)                       // trim input

		// menu switch
		if strings.HasPrefix(input, "\\") {
			if strings.HasPrefix(input, "\\\\") {
				// omok move command
				if end_game {
					fmt.Println("game already finished.")
					time.Sleep(1 * time.Second)
					continue
				}
				if !my_turn {
					fmt.Println("not your turn.")
					time.Sleep(1 * time.Second)
					continue
				}
				if len(input) < 6 {
					fmt.Println("invalid input.")
					time.Sleep(1 * time.Second)
					continue
				}
				input = input[3:]
				xy := strings.Split(input, " ")
				if len(xy) != 2 {
					fmt.Println("invalid input.")
					time.Sleep(1 * time.Second)
					continue
				}
				x, err1 := strconv.Atoi(xy[0])
				y, err2 := strconv.Atoi(xy[1])
				if err1 != nil || err2 != nil {
					fmt.Println("invalid input.")
					time.Sleep(1 * time.Second)
					continue
				}
				if x < 0 || y < 0 || x >= Row || y >= Col || board[x][y] != 0 {
					fmt.Println("invalid move")
					time.Sleep(1 * time.Second)
					continue
				}
				msg := "1" + strconv.Itoa(x) + "_" + strconv.Itoa(y) + "|"
				pconn.WriteTo([]byte(msg), opp_address)
				board[x][y] = turn
				printBoard(board)
				my_turn = false
				turn_count += 1

				// check win
				win = checkWin(board, x, y)
				if win != 0 {
					fmt.Printf("you win!")
					msg := "4|"
					pconn.WriteTo([]byte(msg), opp_address)
					end_game = true
				}

				// check draw
				if turn_count == Row*Col {
					fmt.Println("draw!")
					msg := "5|"
					pconn.WriteTo([]byte(msg), opp_address)
					end_game = true
				}

			} else if input == "\\gg" {
				if end_game {
					fmt.Println("game already finished.")
					continue
				}
				fmt.Println("you lose")
				msg := "2|"
				pconn.WriteTo([]byte(msg), opp_address)
				end_game = true
				my_turn = false
			} else if input == "\\exit" {
				if !end_game {
					fmt.Println("you lose")
				}
				msg := "3|"
				pconn.WriteTo([]byte(msg), opp_address)
				return 1
			} else {
				fmt.Println("invalid command.")
				continue
			}
		} else {
			msg := "0" + input + "|"
			pconn.WriteTo([]byte(msg), opp_address)
		}
	}
}

func handleRecMsg(pconn net.PacketConn, opp_name string, opp_address net.Addr) {
	/**
	* msg
	* 1. "0[msg]|" : just chatting
	* 2. "1[num_num]|" : omok
	* 3. "2|" : gg
	* 4. "3|" : exit
	* 5. "4|" : opp win
	* 6. "5|" : draw
	* 7. "6|" : timeout
	**/
	for {
		msg_from_opp := ""
		for {
			buffer := make([]byte, 1024)
			count, _, err := pconn.ReadFrom(buffer)
			if err != nil {
				ch <- 1
				break
			}
			msg := string(buffer)[:count]
			msg_from_opp += msg
			if strings.HasSuffix(msg, "|") {
				msg_from_opp = msg_from_opp[:len(msg_from_opp)-1]
				break
			}
		}
		//elapsedTime := time.Since(timeStart).Seconds() * 1000 // time after get msg

		if strings.HasPrefix(msg_from_opp, "0") {
			// just chatting
			fmt.Println(opp_name + "> " + msg_from_opp[1:])
		} else if strings.HasPrefix(msg_from_opp, "1") {
			// omok
			val := strings.Split(msg_from_opp[1:], "_")
			x, _ := strconv.Atoi(val[0])
			y, _ := strconv.Atoi(val[1])
			my_turn = true

			// update omok
			if turn == 1 {
				board[x][y] = 2
			} else {
				board[x][y] = 1
			}

			// print omok
			fmt.Println()
			printBoard(board)
			turn_count += 1

			// 10 sec timer
			go func(turn_count_before int) {
				<-time.After(time.Second * 10)
				if !end_game && turn_count_before == turn_count {
					fmt.Println("time out! \n you lose.")
					end_game = true
					my_turn = false
					msg := "6|"
					pconn.WriteTo([]byte(msg), opp_address)
				}
			}(turn_count)

		} else if strings.HasPrefix(msg_from_opp, "2") {
			// gg -> I win
			fmt.Println(opp_name + " gives up.\n you win!")
			my_turn = false
			end_game = true
		} else if strings.HasPrefix(msg_from_opp, "4") {
			// opp win,
			fmt.Println("you lose TT")
			my_turn = false
			end_game = true
		} else if strings.HasPrefix(msg_from_opp, "5") {
			// draw!
			fmt.Println("draw!!")
			my_turn = false
			end_game = true
		} else if strings.HasPrefix(msg_from_opp, "6") {
			// timeout
			fmt.Println(opp_name + " timed out.\n you win!")
			my_turn = false
			end_game = true
		} else {
			// exit
			fmt.Println(opp_name + " exited the game.")
			if end_game == false {
				fmt.Println("you win!")
				end_game = true
			}
		}
	}
}

func firstTCP(client_name string, client_udp_port int, pconn net.PacketConn, conn net.Conn) (string, string, string, int) {
	// TCP connection

	fmt.Println("welcome " + client_name + " to p2p-omok server at " + SERVER_IP + ":" + SERVER_PORT + ".")
	// send my info to server
	msg := "," + client_name + "," + strconv.Itoa(client_udp_port)
	conn.Write([]byte(msg))

	// get msg from server
	for {
		msg_from_c := ""
		for {
			buffer := make([]byte, 1024)
			count, _ := conn.Read(buffer)
			msg := string(buffer)[:count]
			msg_from_c += msg
			if strings.HasSuffix(msg, "|") {
				break
			}
		}
		if strings.HasPrefix(msg_from_c, "!wait") { // have to wait
			fmt.Println("waiting for an opponent")
			fmt.Println()
			fmt.Println()
			continue
		} else {
			// get opponent, disconnect tcp
			tcp_response := strings.Split(string(msg_from_c), ",")
			opp_name := tcp_response[1]
			opp_UDP_port := tcp_response[2]
			opp_IP := tcp_response[3]
			_turn := tcp_response[4][:1]
			turn := 1
			opp_address, _ := net.ResolveUDPAddr("udp", opp_IP+":"+opp_UDP_port)

			if strings.Compare(_turn, "2") == 0 { // me first
				fmt.Println(opp_name + " joined (" + opp_IP + ":" + opp_UDP_port + ").")
				fmt.Println(" you play first.")
				turn = 1
				my_turn = true
				turn_count = 0

				// 10 sec timer
				go func(turn_count_before int) {
					<-time.After(time.Second * 10)
					if !end_game && turn_count_before == turn_count {
						fmt.Println("time out! \n you lose.")
						end_game = true
						my_turn = false
						msg := "6|"
						pconn.WriteTo([]byte(msg), opp_address)
					}
				}(turn_count)
			} else { // opp first
				fmt.Println(opp_name + " is waiting for you (" + opp_IP + ":" + opp_UDP_port + "). ")
				fmt.Println(opp_name + " play first.")
				turn = 2
			}

			conn.Close()
			which_conn = "pconn"
			return opp_name, opp_UDP_port, opp_IP, turn
		}
	}
}

// handle ctrl+c
func SetupCloseHandler1(conn net.Conn) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Print()
		if which_conn == "conn" {
			fmt.Printf("Bye~ \n")
			conn.Write([]byte("exit" + client_name))
			conn.Close()
			os.Exit(1)
		}
	}()
}

// handle ctrl+c
func SetupCloseHandler2(pconn net.PacketConn, opp_address net.Addr) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Print()
		if which_conn == "pconn" {
			fmt.Printf("Bye~ \n")
			pconn.WriteTo([]byte("3|"), opp_address)
			pconn.Close()
			os.Exit(0)
		}
	}()
}

// -------------------------------------------------------------------------- Omok util
func printBoard(b Board) {
	fmt.Print("   ")
	for j := 0; j < Col; j++ {
		fmt.Printf("%2d", j)
	}

	fmt.Println()
	fmt.Print("  ")
	for j := 0; j < 2*Col+3; j++ {
		fmt.Print("-")
	}

	fmt.Println()

	for i := 0; i < Row; i++ {
		fmt.Printf("%d |", i)
		for j := 0; j < Col; j++ {
			c := b[i][j]
			if c == 0 {
				fmt.Print(" +")
			} else if c == 1 {
				fmt.Print(" 0")
			} else if c == 2 {
				fmt.Print(" @")
			} else {
				fmt.Print(" |")
			}
		}

		fmt.Println(" |")
	}

	fmt.Print("  ")
	for j := 0; j < 2*Col+3; j++ {
		fmt.Print("-")
	}

	fmt.Println()
}

func checkWin(b Board, x, y int) int {
	lastStone := b[x][y]
	startX, startY, endX, endY := x, y, x, y

	// Check X
	for startX-1 >= 0 && b[startX-1][y] == lastStone {
		startX--
	}
	for endX+1 < Row && b[endX+1][y] == lastStone {
		endX++
	}

	if endX-startX+1 >= 5 {
		return lastStone
	}

	// Check Y
	startX, startY, endX, endY = x, y, x, y
	for startY-1 >= 0 && b[x][startY-1] == lastStone {
		startY--
	}
	for endY+1 < Row && b[x][endY+1] == lastStone {
		endY++
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	// Check Diag 1
	startX, startY, endX, endY = x, y, x, y
	for startX-1 >= 0 && startY-1 >= 0 && b[startX-1][startY-1] == lastStone {
		startX--
		startY--
	}
	for endX+1 < Row && endY+1 < Col && b[endX+1][endY+1] == lastStone {
		endX++
		endY++
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	// Check Diag 2
	startX, startY, endX, endY = x, y, x, y
	for startX-1 >= 0 && endY+1 < Col && b[startX-1][endY+1] == lastStone {
		startX--
		endY++
	}
	for endX+1 < Row && startY-1 >= 0 && b[endX+1][startY-1] == lastStone {
		endX++
		startY--
	}

	if endY-startY+1 >= 5 {
		return lastStone
	}

	return 0
}

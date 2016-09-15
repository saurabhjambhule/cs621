package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {

	// listen on all interfaces
	listener, err := net.Listen("tcp", "127.0.0.1:8081")
	checkError(err)

	for {

	player_1:
		// accept player 1 connection request
		conn_1, err_1 := listener.Accept()
		if err_1 != nil {
			goto player_1
		}
		id_1, err_11 := getPlayerId(conn_1)
		if err_11 {
			goto player_1
		}
		fmt.Println("Connected player-1: %s (%s)", id_1, conn_1.RemoteAddr())

	player_2:
		// accept player 1 connection request
		conn_2, err_2 := listener.Accept()
		if err_2 != nil {
			goto player_2
		}
		id_2, err_22 := getPlayerId(conn_2)
		if err_22 {
			goto player_2
		}
		fmt.Println("Connected player-2: %s (%s)", id_2, conn_2.RemoteAddr())

		/* create database file to record game
		   db, err := os.Create(id_1+"_"+id_2+".txt")
		   if err != nil {
		       fmt.Fprintf(os.Stderr, "Fatal error (%s - %s): %s", id_1, id_2, err.Error())
		       conn_1.Close()
		       conn_2.Close()
		   }
		*/

		// run as a goroutine
		go startGame(conn_1, conn_2, id_1, id_2)
	}
}

func startGame(conn_1 net.Conn, conn_2 net.Conn, id_1 string, id_2 string) {
	// close connections on exit
	defer func() {
		conn_1.Close()
		conn_2.Close()
	}()

	// initialize game board (1-white card, 0-black card)
	gameBoard := [8][8]int{}
	gameBoard[3][3] = 1
	gameBoard[3][4] = 0
	gameBoard[4][3] = 0
	gameBoard[4][4] = 1

	// send start signal to players
	conn_1.Write([]byte("START WHITE" + "\n"))
	conn_2.Write([]byte("START BLACK" + "\n"))

	for {
		//------------| Player_1 Turn |------------>>
		result := isPossibleMove(gameBoard)
		if !result {
			announceResult(conn_2, conn_1, id_1, id_2, id_2, gameBoard, "No more move left to play")
		}

		// will listen for message to process ending in newline (\n) from player_1
		move_1, err_1 := bufio.NewReader(conn_1).ReadString('\n')
		if err_1 != nil {
			return
		}

		coordinates := strings.Split(move_1, " ")
		xCo, _ := strconv.Atoi(coordinates[0])
		yCo, _ := strconv.Atoi(coordinates[1])
		result = checkLegality(gameBoard, xCo, yCo)
		if !result {
			announceResult(conn_2, conn_1, id_1, id_2, id_2, gameBoard, "Illegal move played")
		}
		// update gameboard with player_1 move
		gameBoard[xCo][yCo] = 1
		// send player_1's move to player_2
		conn_2.Write([]byte(move_1 + "\n"))

		//------------| Player_2 Turn |------------>>
		result = isPossibleMove(gameBoard)
		if !result {
			announceResult(conn_1, conn_2, id_1, id_2, id_1, gameBoard, "No more move left to play")
		}
		// will listen for message to process ending in newline (\n) from player_2
		move_2, err_2 := bufio.NewReader(conn_2).ReadString('\n')
		if err_2 != nil {
			return
		}

		coordinates = strings.Split(move_2, " ")
		xCo, _ = strconv.Atoi(coordinates[0])
		yCo, _ = strconv.Atoi(coordinates[1])
		result = checkLegality(gameBoard, xCo, yCo)
		if !result {
			announceResult(conn_1, conn_2, id_1, id_2, id_1, gameBoard, "Illegal move played")
		}

		gameBoard[xCo][yCo] = 0
		// send player_2's move to player_1
		conn_1.Write([]byte(move_2 + "\n"))
	}
}

// get player's Id
func getPlayerId(conn net.Conn) (string, bool) {
	id, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		conn.Write([]byte("Error! No Id.\n"))
		conn.Close()
		return "", true
	}
	return id, false
}

// fuction to check legality of player's move
func checkLegality(gameBoard [8][8]int, xCo int, yCo int) bool {

	return true
}

// fuction to check possibity to do move
func isPossibleMove(gameBoard [8][8]int) bool {

	return true
}

// announce game result and store to database
func announceResult(winner net.Conn, loser net.Conn, id_1 string, id_2 string, win string, gameBoard [8][8]int, message string) {
	fmt.Println("--------------------------")
	fmt.Println("Player-1: %s", id_1)
	fmt.Println("Player-2: %s", id_2)
	fmt.Println("> Winner : %s", win)
	fmt.Println(gameBoard)
	fmt.Println("--------------------------")
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

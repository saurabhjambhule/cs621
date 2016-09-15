package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

func main() {

	// listen on all interfaces
	listener, err := net.Listen("tcp", "127.0.0.1:8081")
	checkError(err)

	// create log file to record game statistics
	log, err := os.Create("result_log.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error %s", err.Error())
	}
	defer log.Close()

	logger := csv.NewWriter(log)

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

		// run as a goroutine
		go startGame(conn_1, conn_2, id_1, id_2, logger)
	}
}

func startGame(conn_1 net.Conn, conn_2 net.Conn, id_1 string, id_2 string, logger *csv.Writer) {
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

	// initialize counter for time average
	var timeCount_1 time.Duration
	var timeCount_2 time.Duration
	var moveCount_1 time.Duration
	var moveCount_2 time.Duration

	// send start signal to players
	conn_1.Write([]byte("START WHITE\n"))
	conn_2.Write([]byte("START BLACK\n"))

	for {
		// game timer start
		gameNow := time.Now()

		//------------| Player_1 Turn |------------>>
		result := isMovePossible(gameBoard, 1)
		if !result {
			gameThen := time.Now()
			gameTime := gameThen.Sub(gameNow)
			announceResult(logger, conn_1, conn_2, id_1, id_2, (timeCount_1 / moveCount_1), (timeCount_2 / moveCount_2), "player_2", gameBoard, "No more move left to play", gameTime)
			return
		}

		// will listen for message to process ending in newline (\n) from player_1
		now := time.Now()
		move_1, err_1 := bufio.NewReader(conn_1).ReadString('\n')
		if err_1 != nil {
			return
		}
		then := time.Now()
		diff := then.Sub(now)
		timeCount_1 = timeCount_1 + diff

		coordinates := strings.Split(move_1, " ")
		xCo, _ := strconv.Atoi(coordinates[0])
		yCo, _ := strconv.Atoi(coordinates[1])
		result = checkLegality(&gameBoard, xCo, yCo, 1)
		if !result {
			gameThen := time.Now()
			gameTime := gameThen.Sub(gameNow)
			announceResult(logger, conn_1, conn_2, id_1, id_2, (timeCount_1 / moveCount_1), (timeCount_2 / moveCount_2), "player_2", gameBoard, "Illegal move played", gameTime)
			return
		}
		// update gameboard with player_1 move
		gameBoard[xCo][yCo] = 1
		// send player_1's move to player_2
		conn_2.Write([]byte(move_1 + "\n"))

		//------------| Player_2 Turn |------------>>
		result = isMovePossible(gameBoard, 2)
		if !result {
			gameThen := time.Now()
			gameTime := gameThen.Sub(gameNow)
			announceResult(logger, conn_1, conn_2, id_1, id_2, (timeCount_1 / moveCount_1), (timeCount_2 / moveCount_2), "player_1", gameBoard, "No more move left to play", gameTime)
			return
		}
		// will listen for message to process ending in newline (\n) from player_2
		now = time.Now()
		move_2, err_2 := bufio.NewReader(conn_2).ReadString('\n')
		if err_2 != nil {
			return
		}
		then = time.Now()
		diff = then.Sub(now)
		timeCount_2 = timeCount_2 + diff

		coordinates = strings.Split(move_2, " ")
		xCo, _ = strconv.Atoi(coordinates[0])
		yCo, _ = strconv.Atoi(coordinates[1])
		result = checkLegality(&gameBoard, xCo, yCo, 2)
		if !result {
			gameThen := time.Now()
			gameTime := gameThen.Sub(gameNow)
			announceResult(logger, conn_1, conn_2, id_1, id_2, (timeCount_1 / moveCount_1), (timeCount_2 / moveCount_2), "player_1", gameBoard, "Illegal move played", gameTime)
			return
		}

		gameBoard[xCo][yCo] = 2
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

	rollNo := strings.Split(id, " ")
	if len(rollNo[0]) < 9 {
		conn.Write([]byte("Error! No Id.\n"))
		conn.Close()
		return "", true
	}
	return rollNo[0], false
}

// fuction to check legality of player's move
func checkLegality(gameBoard *[8][8]int, xCo int, yCo int, player int) bool {
	mapPlayer := make(map[int]int)
	mapPlayer[1] = 2
	mapPlayer[2] = 1
	flag := false

	// check in East direction
	first := true
	found := false
	upto := yCo
	for i := yCo + 1; i < 8; i++ {
		if first {
			first = false
			if gameBoard[xCo][i] != mapPlayer[player] {
				break
			}
		} else {
			if gameBoard[xCo][i] == 0 {
				break
			} else if gameBoard[xCo][i] == player {
				found = true
				upto = i
				break
			}
		}
	}
	if found {
		flag = true
		for i := yCo + 1; i < upto; i++ {
			gameBoard[xCo][i] = player
		}
	}

	// check in West direction
	first = true
	found = false
	upto = yCo
	for i := yCo - 1; i >= 0; i-- {
		if first {
			first = false
			if gameBoard[xCo][i] != mapPlayer[player] {
				break
			}
		} else {
			if gameBoard[xCo][i] == 0 {
				break
			} else if gameBoard[xCo][i] == player {
				found = true
				upto = i
				break
			}
		}
	}
	if found {
		flag = true
		for i := yCo - 1; i > upto; i-- {
			gameBoard[xCo][i] = player
		}
	}

	// check in North direction
	first = true
	found = false
	upto = xCo
	for i := xCo - 1; i >= 0; i-- {
		if first {
			first = false
			if gameBoard[i][yCo] != mapPlayer[player] {
				break
			}
		} else {
			if gameBoard[xCo][i] == 0 {
				break
			} else if gameBoard[i][yCo] == player {
				found = true
				upto = i
				break
			}
		}
	}
	if found {
		flag = true
		for i := xCo - 1; i > upto; i-- {
			gameBoard[i][yCo] = player
		}
	}

	// check in South direction
	first = true
	found = false
	upto = xCo
	for i := xCo + 1; i < 8; i++ {
		if first {
			first = false
			if gameBoard[i][yCo] != mapPlayer[player] {
				break
			}
		} else {
			if gameBoard[xCo][i] == 0 {
				break
			} else if gameBoard[i][yCo] == player {
				found = true
				upto = i
				break
			}
		}
	}
	if found {
		flag = true
		for i := xCo + 1; i < upto; i++ {
			gameBoard[i][yCo] = player
		}
	}

	// check in North-West direction
	first = true
	found = false
	uptoX := xCo
	uptoY := yCo
	i := xCo - 1
	j := yCo - 1
	for i >= 0 && j >= 0 {
		if first {
			first = false
			if gameBoard[i][j] != mapPlayer[player] {
				break
			}
		} else {
			if gameBoard[i][j] == 0 {
				break
			} else if gameBoard[i][j] == player {
				found = true
				uptoX = i
				uptoY = j
				break
			}
		}
		i--
		j--
	}
	if found {
		flag = true
		i = xCo - 1
		j = yCo - 1
		for i > uptoX && j > uptoY {
			gameBoard[i][j] = player
			i--
			j--
		}
	}

	// check in South-East direction
	first = true
	found = false
	uptoX = xCo
	uptoY = yCo
	i = xCo + 1
	j = yCo + 1
	for i < 8 && j < 8 {
		if first {
			first = false
			if gameBoard[i][j] != mapPlayer[player] {
				break
			}
		} else {
			if gameBoard[i][j] == 0 {
				break
			} else if gameBoard[i][j] == player {
				found = true
				uptoX = i
				uptoY = j
				break
			}
		}
		i++
		j++
	}
	if found {
		flag = true
		i = xCo + 1
		j = yCo + 1
		for i < uptoX && j < uptoY {
			gameBoard[i][j] = player
			i++
			j++
		}
	}

	// check in North-East direction
	first = true
	found = false
	uptoX = xCo
	uptoY = yCo
	i = xCo - 1
	j = yCo + 1
	for i >= 0 && j < 8 {
		if first {
			first = false
			if gameBoard[i][j] != mapPlayer[player] {
				break
			}
		} else {
			if gameBoard[i][j] == 0 {
				break
			} else if gameBoard[i][j] == player {
				found = true
				uptoX = i
				uptoY = j
				break
			}
		}
		i--
		j++
	}
	if found {
		flag = true
		i = xCo - 1
		j = yCo + 1
		for i > uptoX && j < uptoY {
			gameBoard[i][j] = player
			i--
			j++
		}
	}

	// check in South-West direction
	first = true
	found = false
	uptoX = xCo
	uptoY = yCo
	i = xCo + 1
	j = yCo - 1
	for i < 8 && j >= 0 {
		if first {
			first = false
			if gameBoard[i][j] != mapPlayer[player] {
				break
			}
		} else {
			if gameBoard[i][j] == 0 {
				break
			} else if gameBoard[i][j] == player {
				found = true
				uptoX = i
				uptoY = j
				break
			}
		}
		i++
		j--
	}
	if found {
		flag = true
		i = xCo + 1
		j = yCo - 1
		for i < uptoX && j > uptoY {
			gameBoard[i][j] = player
			i++
			j--
		}
	}

	return flag
}

// fuction to check possibity to do move
func isMovePossible(gameBoard [8][8]int, player int) bool {

	return true
}

// announce game result and store to database
func announceResult(logger *csv.Writer, conn_1 net.Conn, conn_2 net.Conn, id_1 string, id_2 string, avg_1 time.Duration, avg_2 time.Duration, win string, gameBoard [8][8]int, message string, gameTime time.Duration) {

	// calculate game points of each player
	count_1 := 0
	count_2 := 0
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			if gameBoard[i][j] == 1 {
				count_1++
			} else if gameBoard[i][j] == 2 {
				count_2++
			}
		}
	}

	// write down game stats in log file
	var log1 = []string{id_1, strconv.Itoa(count_1), avg_1.String()}
	var log2 = []string{id_2, strconv.Itoa(count_2), avg_2.String()}
	var log3 = []string{win, message, gameTime.String()}
	var log4 = []string{" ", " ", " "}
	mutex.Lock()
	logger.Write(log1)
	logger.Write(log2)
	logger.Write(log3)
	logger.Write(log4)
	mutex.Unlock()

	fmt.Println("--------------------------")
	fmt.Println("Player-1: %s", id_1)
	fmt.Println("Player-2: %s", id_2)
	if win == "player_1" {
		fmt.Println("> Winner : %s (Score= %d)", id_1, count_1)
		fmt.Println("> Looser : %s (Score= %d)", id_2, count_2)
		conn_1.Write([]byte("Congrats! You won. (Score= " + strconv.Itoa(count_1) + ")\n"))
		conn_1.Write([]byte("Sorry! You lose. (Score= " + strconv.Itoa(count_2) + ")\n"))
	} else {
		fmt.Println("> Winner : %s (Score= %d)", id_2, count_2)
		fmt.Println("> Looser : %s (Score= %d)", id_1, count_1)
		conn_2.Write([]byte("Congrats! You won. (Score= " + strconv.Itoa(count_2) + ")\n"))
		conn_1.Write([]byte("Sorry! You lose. (Score= " + strconv.Itoa(count_1) + ")\n"))
	}

	fmt.Println(gameBoard)
	fmt.Println("--------------------------")
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

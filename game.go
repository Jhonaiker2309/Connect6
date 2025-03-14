package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const SIZE = 19

type Board [SIZE][SIZE]rune

func main() {
	var board Board
	initializeBoard(&board)
	reader := bufio.NewReader(os.Stdin)
	currentPlayer := 'X'
	stonesToPlace := 1

	for {
		printBoard(board)
		fmt.Printf("Jugador %c: Coloca %d piedra(s)\n", currentPlayer, stonesToPlace)

		for i := 0; i < stonesToPlace; i++ {
			for {
				fmt.Printf("Piedra %d (fila columna 1-%d): ", i+1, SIZE)
				input, _ := reader.ReadString('\n')
				coords := strings.Fields(strings.TrimSpace(input))

				if len(coords) != 2 {
					fmt.Println("Entrada inválida. Usa: fila columna")
					continue
				}

				row, err1 := strconv.Atoi(coords[0])
				col, err2 := strconv.Atoi(coords[1])

				if err1 != nil || err2 != nil {
					fmt.Println("Deben ser números!")
					continue
				}

				row-- // Convertir a índice 0-based
				col-- 

				if row < 0 || row >= SIZE || col < 0 || col >= SIZE {
					fmt.Println("Posición fuera del tablero!")
					continue
				}

				if board[row][col] != '.' {
					fmt.Println("Casilla ocupada!")
					continue
				}

				board[row][col] = currentPlayer
				break
			}
		}

		if currentPlayer == 'X' {
			currentPlayer = 'O'
		} else {
			currentPlayer = 'X'
		}
		stonesToPlace = 2
	}
}

func initializeBoard(b *Board) {
	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			b[i][j] = '.'
		}
	}
}

func printBoard(b Board) {
	fmt.Print("   ")
	for i := 0; i < SIZE; i++ {
		fmt.Printf("%2d ", i+1)
	}
	fmt.Println()

	for i := 0; i < SIZE; i++ {
		fmt.Printf("%2d ", i+1)
		for j := 0; j < SIZE; j++ {
			fmt.Printf(" %c ", b[i][j])
		}
		fmt.Println()
	}
}
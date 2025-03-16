package ui

import (
	"fmt"
	"connect6/board"
)

// PrintBoard muestra el tablero con formato legible en consola
// Parámetros:
// - b: Tablero a mostrar
// Formato:
//   - Encabezado con números de columna (0-18)
//   - Filas numeradas (0-18) a la izquierda
//   - 'B' para fichas negras, 'W' para blancas, '.' para celdas vacías
func PrintBoard(b board.Board) {
    fmt.Print("    ") // Ajustar espacio para el encabezado de columnas
    for c := 0; c < board.BoardSize; c++ {
        fmt.Printf("%2d ", c) // Encabezado de columnas (0-18)
    }
    fmt.Println()

    for r := 0; r < board.BoardSize; r++ {
        fmt.Printf("%2d ", r) // Encabezado de filas (0-18)
        for c := 0; c < board.BoardSize; c++ {
            char := b[r][c]
            if char == 0 { // 0 representa celda vacía
                char = '.'
            }
            fmt.Printf(" %c ", char)
        }
        fmt.Println()
    }
    fmt.Println()
}

// GetPlayerMove obtiene y valida el movimiento del jugador humano
// Flujo:
//   1. Solicita entrada con 4 números (fila1 col1 fila2 col2)
//   2. Valida formato numérico
//   3. Valida posiciones con board.IsValidMove
//   4. Repite hasta obtener entrada válida
// Retorna:
//   - Move válido listo para aplicar al tablero
func GetPlayerMove(b board.Board) board.Move {
	var row1, col1, row2, col2 int
	for {
		fmt.Print("Ingresa dos posiciones (fila1 columna1 fila2 columna2): ")
		_, err := fmt.Scan(&row1, &col1, &row2, &col2)
		
		if err != nil {
			fmt.Println("Error: Entrada inválida. Usa 4 números separados por espacios.")
			// Limpiar buffer de entrada
			var discard string
			fmt.Scanln(&discard)
			continue
		}
		
		p1 := board.Position{Row: row1, Col: col1}
		p2 := board.Position{Row: row2, Col: col2}
		
		if board.IsValidMove(b, p1, p2) {
			return board.Move{p1, p2}
		}
		
		fmt.Println("Movimiento inválido. Intenta nuevamente.")
	}
}

// ShowGameMenu muestra el menú de inicio del juego
// Retorna:
//   - 'W' si el jugador elige empezar (s/S)
//   - 'B' para cualquier otra entrada (bot primero)
// Interacción:
//   - Muestra prompt y lee entrada simple
//   - No distingue mayúsculas/minúsculas
func ShowGameMenu() rune {
	var choice string
	fmt.Print("¿Quieres jugar primero? (s/n): ")
	fmt.Scan(&choice)
	if choice == "s" || choice == "S" {
		return 'W' // Jugador es blancas
	}
	return 'B' // Bot es negras
}

// ShowResult muestra el resultado final del juego
// Parámetro:
//   - winner: 'B' (Negras ganan), 'W' (Blancas ganan), ' ' (Empate)
func ShowResult(winner rune) {
	switch winner {
	case 'B':
		fmt.Println("¡Las fichas Negras ganan!")
	case 'W':
		fmt.Println("¡Las fichas Blancas ganan!")
	default:
		fmt.Println("¡Es un empate!")
	}
}
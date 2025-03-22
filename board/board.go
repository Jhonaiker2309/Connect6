package board

import (
	"strings"
)

const (
	BoardSize = 19
	WinLength = 6
)

// Position representa una coordenada en el tablero
// Row: Fila (0-18)
// Col: Columna (0-18)
type Position struct{ Row, Col int }

// Move representa un movimiento con dos posiciones
// [0]: Primera posición
// [1]: Segunda posición
type Move [2]Position

// Board representa el tablero del juego
// Usa '\x00' para celdas vacías, 'B' para negras, 'W' para blancas
type Board [BoardSize][BoardSize]rune

// ApplyMove coloca dos piedras en el tablero
// Parámetros:
// - b: Puntero al tablero
// - move: Movimiento a realizar
// - player: Jugador actual ('B' o 'W')
func ApplyMove(b *Board, move Move, player rune) {
	b[move[0].Row][move[0].Col] = player
	b[move[1].Row][move[1].Col] = player
}

// CheckWin verifica si un jugador ha ganado
// Parámetros:
// - b: Tablero actual
// - player: Jugador a verificar
// Retorna: true si el jugador tiene 6 en línea
func CheckWin(b Board, player rune) bool {
	directions := []struct{ dr, dc int }{
		{0, 1},  // Horizontal
		{1, 0},  // Vertical
		{1, 1},  // Diagonal derecha
		{1, -1}, // Diagonal izquierda
	}

	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] != player {
				continue
			}

			for _, dir := range directions {
				count := 1
				for step := 1; step < WinLength; step++ {
					nr, nc := r+dir.dr*step, c+dir.dc*step
					if nr < 0 || nr >= BoardSize || nc < 0 || nc >= BoardSize {
						break
					}
					if b[nr][nc] != player {
						break
					}
					count++
				}
				if count >= WinLength {
					return true
				}
			}
		}
	}
	return false
}

// IsValidMove valida si un movimiento es legal
// Parámetros:
// - b: Tablero actual
// - p1: Primera posición del movimiento
// - p2: Segunda posición del movimiento
// Retorna: true si ambas posiciones están vacías y dentro del tablero
func IsValidMove(b Board, p1, p2 Position) bool {
	if p1 == p2 {
		return false
	}

	inRange := func(p Position) bool {
		return p.Row >= 0 && p.Row < BoardSize && p.Col >= 0 && p.Col < BoardSize
	}

	return inRange(p1) && inRange(p2) &&
		b[p1.Row][p1.Col] == '\x00' &&
		b[p2.Row][p2.Col] == '\x00'
}

// SwitchPlayer alterna entre jugadores
// Parámetros:
// - player: Jugador actual
// Retorna: 'W' si input es 'B', 'B' si input es 'W'
func SwitchPlayer(player rune) rune {
	if player == 'B' {
		return 'W'
	}
	return 'B'
}

// CloneBoard crea una copia exacta del tablero
// Parámetros:
// - b: Tablero original
// Retorna: Nueva instancia del tablero
func CloneBoard(b Board) Board {
	var newBoard Board
	for i := range b {
		newBoard[i] = b[i]
	}
	return newBoard
}

// GetCurrentPlayer determina quién debe jugar
// Parámetros:
// - b: Tablero actual
// Retorna: 'B' si igual número de piedras, 'W' si hay más negras
func GetCurrentPlayer(b Board) rune {
	var black, white int
	for _, row := range b {
		for _, cell := range row {
			switch cell {
			case 'B':
				black++
			case 'W':
				white++
			}
		}
	}
	if black == white {
		return 'B'
	}
	return 'W'
}

// EvaluateBoard evalúa la ventaja global de 'player' en el tablero 'b'.
// Retorna un valor positivo si es mejor para 'player', negativo si es mejor para el rival.
func EvaluateBoard(b Board, player rune) float64 {
	opponent := SwitchPlayer(player)
	playerScore := 0
	oppScore := 0

	// Recorremos el tablero en busca de cadenas de 'B' o 'W'.
	// Para cada casilla con una ficha, examinamos la longitud de la cadena en
	// 4 direcciones. Obtenemos la longitud y si está bloqueada a izquierda/derecha.
	directions := []struct{ dr, dc int }{
		{0, 1},  // Horizontal
		{1, 0},  // Vertical
		{1, 1},  // Diagonal \
		{1, -1}, // Diagonal /
	}

	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			cell := b[r][c]
			if cell == '\x00' {
				continue
			}
			// Vemos para cada dirección
			for _, d := range directions {
				length, blockedA, blockedB := chainInfo(b, r, c, d.dr, d.dc, cell)

				if cell == player {
					playerScore += WeightedChainScore(length, blockedA, blockedB)
				} else if cell == opponent {
					oppScore += WeightedChainScore(length, blockedA, blockedB)
				}
			}
		}
	}

	// Un valor final. Podríamos normalizarlo, pero por simplicidad
	// devolvemos la diferencia. Cuanto mayor => más favorable a 'player'.
	return float64(playerScore - oppScore)
}

// WeightedChainScore asigna un valor según la longitud de la cadena y si
// está bloqueada a uno o ambos extremos.
func WeightedChainScore(length int, blockedA, blockedB bool) int {
	// Puntos base si la cadena es >= 6 => victoria instantánea
	if length >= 6 {
		return 999999
	}

	openEnds := 0
	if !blockedA {
		openEnds++
	}
	if !blockedB {
		openEnds++
	}

	// Ejemplo de escalado
	// 5 en línea:
	if length == 5 {
		// con 2 extremos abiertos => un movimiento (2 piedras) => gana
		if openEnds == 2 {
			return 100000
		} else if openEnds == 1 {
			return 50000
		}
		return 20000
	}

	// 4 en línea:
	// Si openEnds=2 => se puede poner 2 fichas y hacer 6
	// p.ej. 4 + 2 = 6 => su peligrosidad sube mucho
	if length == 4 {
		if openEnds == 2 {
			return 30000 // sube este valor
		} else if openEnds == 1 {
			return 15000
		}
		return 5000
	}

	// 3 en línea:
	if length == 3 {
		if openEnds == 2 {
			// 3 + 2 = 5 => no gana de inmediato, pero se queda a 1
			// Súbelo un poco
			return 7000
		} else if openEnds == 1 {
			return 3000
		}
		return 1000
	}

	// 2 en línea:
	if length == 2 {
		if openEnds == 2 {
			return 1500
		}
		return 500
	}

	// 1 sola
	if length == 1 {
		return 50
	}

	return 0
}

// chainInfo retorna la longitud de la cadena que inicia en (r,c) en la dirección (dr, dc),
// y si está bloqueada en los extremos.
func chainInfo(b Board, r, c, dr, dc int, player rune) (length int, blockedA, blockedB bool) {
	// Empezamos con length=1 (la propia casilla (r,c))
	length = 1

	// Avanzamos "hacia adelante" en la dirección (dr,dc)
	step := 1
	for {
		nr := r + dr*step
		nc := c + dc*step
		if nr < 0 || nr >= BoardSize || nc < 0 || nc >= BoardSize {
			// salimos del tablero => se considera "bloqueado"
			blockedB = true
			break
		}
		if b[nr][nc] == player {
			length++
			step++
			continue
		}
		if b[nr][nc] == '\x00' {
			// hueco => no bloqueado, paramos la cadena
			blockedB = false
		} else {
			// hay una ficha rival => bloqueado
			blockedB = true
		}
		break
	}

	// Avanzamos "hacia atrás" en la dirección opuesta (-dr, -dc)
	step = 1
	for {
		nr := r - dr*step
		nc := c - dc*step
		if nr < 0 || nr >= BoardSize || nc < 0 || nc >= BoardSize {
			blockedA = true
			break
		}
		if b[nr][nc] == player {
			length++
			step++
			continue
		}
		if b[nr][nc] == '\x00' {
			blockedA = false
		} else {
			blockedA = true
		}
		break
	}

	return length, blockedA, blockedB
}

// EvaluatePosition calcula valor estratégico de una posición
// Parámetros:
// - b: Tablero actual
// - r: Fila a evaluar
// - c: Columna a evaluar
// - player: Jugador a evaluar
// Retorna: Puntaje numérico (mayor = mejor posición)
func EvaluatePosition(b Board, r, c int, player rune) int {
	score := 0
	directions := []struct{ dr, dc int }{
		{0, 1}, {1, 0}, {1, 1}, {1, -1},
	}

	for _, dir := range directions {
		streak := 1
		openEnds := 0
		// contamos fichas continuas hacia adelante
		for step := 1; step < WinLength; step++ {
			nr, nc := r+dir.dr*step, c+dir.dc*step
			if nr < 0 || nr >= BoardSize || nc < 0 || nc >= BoardSize {
				break
			}
			if b[nr][nc] == player {
				streak++
			} else if b[nr][nc] == '\x00' {
				openEnds++
				break
			} else {
				// bloqueado por oponente
				break
			}
		}
		// puntuación simple
		score += streak * streak * (openEnds + 1)
	}
	return score
}

// baseSmartMoves genera movimientos "básicos" sin filtrar demasiado
// Retorna: Lista de hasta 100 pares de posiciones prioritarias
func baseSmartMoves(b Board) []Move {
	positions := GetPriorityPositions(b, 2)
	var moves []Move
	maxPairs := 100

	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			moves = append(moves, Move{positions[i], positions[j]})
			if len(moves) >= maxPairs {
				return moves
			}
		}
	}
	return moves
}

// GenerateSmartMoves genera movimientos estratégicos
// En esta versión, añadimos la idea de priorizar ciertas jugadas.
// Ejemplo: si hubiera una jugada ganadora para B o W (aunque sin saber quién juega),
// se agregan primero. Luego se añaden las jugadas base.
func GenerateSmartMoves(b Board) []Move {
	var moves []Move

	// 1) Jugada ganadora para negras
	if winB := FindWinningMove(b, 'B'); winB != nil {
		moves = append(moves, *winB)
	}
	// 2) Jugada ganadora para blancas
	if winW := FindWinningMove(b, 'W'); winW != nil {
		moves = append(moves, *winW)
	}

	// 3) baseSmartMoves
	base := baseSmartMoves(b)
	moves = append(moves, base...)

	// Opcional: recortar
	if len(moves) > 150 {
		moves = moves[:150]
	}
	return moves
}

// FindWinningMove busca victoria inmediata
// Parámetros:
// - b: Tablero actual
// - player: Jugador a verificar
// Retorna: Movimiento ganador si existe, nil en caso contrario
func FindWinningMove(b Board, player rune) *Move {
	moves := baseSmartMoves(b)
	for _, move := range moves {
		testBoard := CloneBoard(b)
		ApplyMove(&testBoard, move, player)
		if CheckWin(testBoard, player) {
			return &move
		}
	}
	return nil
}

func FindPairWinningMove(b Board, player rune) *Move {
	// Generar todos los movimientos posibles para el primer paso
	firstMoves := GenerateSmartMoves(b)

	// Verificar cada par de movimientos consecutivos
	for _, firstMove := range firstMoves {
		// Aplicar primer movimiento
		testBoard := CloneBoard(b)
		ApplyMove(&testBoard, firstMove, player)

		// Generar movimientos para el segundo paso
		secondMoves := GenerateSmartMoves(testBoard)

		for _, secondMove := range secondMoves {
			// Aplicar segundo movimiento
			finalBoard := CloneBoard(testBoard)
			ApplyMove(&finalBoard, secondMove, player)

			// Verificar si se completa la victoria
			if CheckWin(finalBoard, player) {
				return &firstMove // Devolver el primer movimiento del par ganador
			}
		}
	}
	return nil
}

// GetPriorityPositions obtiene ubicaciones clave
// Añade el área central 5x5 si el tablero está vacío, etc.
func GetPriorityPositions(b Board, radius int) []Position {
	positions := make(map[Position]bool)
	center := BoardSize / 2

	if IsBoardEmpty(b) {
		// área central 5x5
		for dr := -2; dr <= 2; dr++ {
			for dc := -2; dc <= 2; dc++ {
				nr, nc := center+dr, center+dc
				if nr >= 0 && nr < BoardSize && nc >= 0 && nc < BoardSize {
					positions[Position{nr, nc}] = true
				}
			}
		}
		return mapToSlice(positions)
	}

	// Posiciones cerca de piedras existentes
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] != '\x00' {
				for dr := -radius; dr <= radius; dr++ {
					for dc := -radius; dc <= radius; dc++ {
						nr, nc := r+dr, c+dc
						if nr >= 0 && nr < BoardSize && nc >= 0 && nc < BoardSize && b[nr][nc] == '\x00' {
							positions[Position{nr, nc}] = true
						}
					}
				}
			}
		}
	}
	return mapToSlice(positions)
}

// IsBoardEmpty verifica si el tablero está vacío
func IsBoardEmpty(b Board) bool {
	for _, row := range b {
		for _, cell := range row {
			if cell != '\x00' {
				return false
			}
		}
	}
	return true
}

// mapToSlice convierte mapa de posiciones a slice
func mapToSlice(m map[Position]bool) []Position {
	result := make([]Position, 0, len(m))
	for pos := range m {
		result = append(result, pos)
	}
	return result
}

// GetWinner determina el ganador del juego
// Retorna: 'B', 'W' o ' ' (sin ganador)
func GetWinner(board Board) rune {
	if CheckWin(board, 'B') {
		return 'B'
	}
	if CheckWin(board, 'W') {
		return 'W'
	}
	return ' '
}

// BoardHash genera un identificador (string) para el estado del tablero
// para usar en tablas de transposición o almacenamiento.
func BoardHash(b Board) string {
	var sb strings.Builder
	sb.Grow(BoardSize * BoardSize)

	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			// Usamos el rune directamente, p.ej. '\x00', 'B', 'W'.
			// '\x00' no es representable textualmente, así que puedes mapearlo a un caracter.
			cell := b[r][c]
			if cell == '\x00' {
				sb.WriteRune('.') // Por ejemplo, '.' para vacío
			} else {
				sb.WriteRune(cell) // 'B' o 'W'
			}
		}
	}
	return sb.String()
}

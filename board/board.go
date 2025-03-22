package board

import (
	"strings"
)

const (
	BoardSize = 19
	WinLength = 6
)

// Position representa una coordenada en el tablero (fila, columna).
type Position struct {
	Row, Col int
}

// Move representa un movimiento, compuesto por dos posiciones.
// Para la primera jugada de un jugador, se utiliza (-1,-1) en move[1] para indicar
// que solo se coloca una ficha.
type Move [2]Position

// Board representa el tablero del juego Connect6.
// Se usa '\x00' para celdas vacías, 'B' para fichas negras y 'W' para fichas blancas.
type Board [BoardSize][BoardSize]rune

// ApplyMove coloca fichas en el tablero.
// Si move[1] es (-1,-1), se coloca solo una ficha; de lo contrario, se colocan dos.
func ApplyMove(b *Board, move Move, player rune) {
	if move[1].Row == -1 && move[1].Col == -1 {
		b[move[0].Row][move[0].Col] = player
	} else {
		b[move[0].Row][move[0].Col] = player
		b[move[1].Row][move[1].Col] = player
	}
}

// CheckWin verifica si un jugador ha ganado (se tienen 6 o más fichas en línea).
func CheckWin(b Board, player rune) bool {
	directions := []struct{ dr, dc int }{
		{0, 1},  // Horizontal
		{1, 0},  // Vertical
		{1, 1},  // Diagonal (de izquierda a derecha)
		{1, -1}, // Diagonal (de derecha a izquierda)
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

// IsValidMove valida que dos posiciones sean diferentes, estén dentro del tablero y vacías.
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

// SwitchPlayer alterna entre jugadores ('B' y 'W').
func SwitchPlayer(player rune) rune {
	if player == 'B' {
		return 'W'
	}
	return 'B'
}

// CloneBoard crea una copia exacta del tablero.
func CloneBoard(b Board) Board {
	var newBoard Board
	for i := range b {
		newBoard[i] = b[i]
	}
	return newBoard
}

// GetCurrentPlayer determina quién debe jugar basado en el número de fichas en el tablero.
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

// EvaluateBoard evalúa la ventaja global para 'player'.
// Retorna un valor positivo si es favorable a 'player', negativo si es favorable al oponente.
func EvaluateBoard(b Board, player rune) float64 {
	opponent := SwitchPlayer(player)
	playerScore := 0
	oppScore := 0

	// Se evalúan las cadenas de fichas en 4 direcciones.
	directions := []struct{ dr, dc int }{
		{0, 1},
		{1, 0},
		{1, 1},
		{1, -1},
	}

	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			cell := b[r][c]
			if cell == '\x00' {
				continue
			}
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
	return float64(playerScore - oppScore)
}

// WeightedChainScore asigna un valor según la longitud de la cadena y si está bloqueada.
func WeightedChainScore(length int, blockedA, blockedB bool) int {
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
	if length == 5 {
		if openEnds == 2 {
			return 100000
		} else if openEnds == 1 {
			return 50000
		}
		return 20000
	}
	if length == 4 {
		if openEnds == 2 {
			return 30000
		} else if openEnds == 1 {
			return 15000
		}
		return 5000
	}
	if length == 3 {
		if openEnds == 2 {
			return 7000
		} else if openEnds == 1 {
			return 3000
		}
		return 1000
	}
	if length == 2 {
		if openEnds == 2 {
			return 1500
		}
		return 500
	}
	if length == 1 {
		return 50
	}
	return 0
}

// chainInfo retorna la longitud de la cadena desde (r,c) en la dirección (dr,dc),
// y si la cadena está bloqueada en cada extremo.
func chainInfo(b Board, r, c, dr, dc int, player rune) (length int, blockedA, blockedB bool) {
	length = 1
	step := 1
	for {
		nr := r + dr*step
		nc := c + dc*step
		if nr < 0 || nr >= BoardSize || nc < 0 || nc >= BoardSize {
			blockedB = true
			break
		}
		if b[nr][nc] == player {
			length++
			step++
			continue
		}
		if b[nr][nc] == '\x00' {
			blockedB = false
		} else {
			blockedB = true
		}
		break
	}
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

// baseSmartMoves genera movimientos "básicos" sin demasiada filtración,
// retornando hasta 100 pares de posiciones prioritarias.
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

// GenerateSmartMoves genera movimientos estratégicos.
// Agrega jugadas ganadoras (para 'B' o 'W') y luego los movimientos básicos.
func GenerateSmartMoves(b Board) []Move {
	var moves []Move
	if winB := FindWinningMove(b, 'B'); winB != nil {
		moves = append(moves, *winB)
	}
	if winW := FindWinningMove(b, 'W'); winW != nil {
		moves = append(moves, *winW)
	}
	base := baseSmartMoves(b)
	moves = append(moves, base...)
	if len(moves) > 150 {
		moves = moves[:150]
	}
	return moves
}

// FindWinningMove busca una jugada que garantice victoria inmediata para 'player'.
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

// FindPairWinningMove busca una pareja de jugadas consecutivas que lleven a la victoria.
func FindPairWinningMove(b Board, player rune) *Move {
	firstMoves := GenerateSmartMoves(b)
	for _, firstMove := range firstMoves {
		testBoard := CloneBoard(b)
		ApplyMove(&testBoard, firstMove, player)
		secondMoves := GenerateSmartMoves(testBoard)
		for _, secondMove := range secondMoves {
			finalBoard := CloneBoard(testBoard)
			ApplyMove(&finalBoard, secondMove, player)
			if CheckWin(finalBoard, player) {
				return &firstMove
			}
		}
	}
	return nil
}

// GetPriorityPositions obtiene posiciones prioritarias.
// Si el tablero está vacío, retorna el área central (ej. 5x5); de lo contrario, retorna posiciones
// cercanas a las fichas existentes.
func GetPriorityPositions(b Board, radius int) []Position {
	positions := make(map[Position]bool)
	center := BoardSize / 2
	if IsBoardEmpty(b) {
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

// IsBoardEmpty verifica si el tablero está completamente vacío.
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

// mapToSlice convierte un mapa de posiciones en un slice.
func mapToSlice(m map[Position]bool) []Position {
	result := make([]Position, 0, len(m))
	for pos := range m {
		result = append(result, pos)
	}
	return result
}

// GetWinner determina el ganador del juego.
// Retorna 'B', 'W' o ' ' si no hay ganador.
func GetWinner(board Board) rune {
	if CheckWin(board, 'B') {
		return 'B'
	}
	if CheckWin(board, 'W') {
		return 'W'
	}
	return ' '
}

// BoardHash genera un identificador para el tablero, útil para tablas de transposición.
func BoardHash(b Board) string {
	var sb strings.Builder
	sb.Grow(BoardSize * BoardSize)
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			cell := b[r][c]
			if cell == '\x00' {
				sb.WriteRune('.') // Se usa '.' para representar vacío
			} else {
				sb.WriteRune(cell)
			}
		}
	}
	return sb.String()
}

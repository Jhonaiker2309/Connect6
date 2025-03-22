package board

// Constantes que definen el tamaño del tablero y la cantidad de fichas en línea necesarias para ganar.
const (
	BoardSize = 19
	WinLength = 6
)

// Position representa una coordenada en el tablero, con fila y columna.
type Position struct {
	Row, Col int
}

// Move define un movimiento en el tablero, compuesto por dos posiciones. En la primera jugada se usa
// la segunda posición con el valor (-1,-1) para indicar que solo se coloca una ficha.
type Move [2]Position

// Board representa el estado del tablero como una matriz de 19x19. Cada celda contiene un rune:
// '\x00' indica celda vacía, 'B' representa fichas negras y 'W' fichas blancas.
type Board [BoardSize][BoardSize]rune

// ApplyMove coloca fichas en el tablero.
// Si move[1] es (-1,-1), se coloca solo una ficha; de lo contrario, se colocan ambas.
func ApplyMove(b *Board, move Move, player rune) {
	if move[1].Row == -1 && move[1].Col == -1 {
		b[move[0].Row][move[0].Col] = player
	} else {
		b[move[0].Row][move[0].Col] = player
		b[move[1].Row][move[1].Col] = player
	}
}

// CheckWin recorre el tablero y verifica si el jugador indicado tiene al menos 6 fichas consecutivas
// en alguna de las cuatro direcciones (horizontal, vertical o en las dos diagonales).
func CheckWin(b Board, player rune) bool {
	directions := []struct{ dr, dc int }{
		{0, 1}, {1, 0}, {1, 1}, {1, -1},
	}
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] != player {
				continue
			}
			for _, d := range directions {
				count := 1
				for step := 1; step < WinLength; step++ {
					nr, nc := r+d.dr*step, c+d.dc*step
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

// IsValidMove determina si dos posiciones (p1 y p2) constituyen un movimiento legal, es decir,
// que sean distintas, estén dentro del tablero y ambas celdas estén vacías.
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

// SwitchPlayer alterna la pieza del jugador. Si el jugador actual es 'B', retorna 'W' y viceversa.
func SwitchPlayer(player rune) rune {
	if player == 'B' {
		return 'W'
	}
	return 'B'
}

// CloneBoard genera una copia exacta del tablero, lo que permite realizar simulaciones sin modificar el estado original.
func CloneBoard(b Board) Board {
	var newBoard Board
	for i := range b {
		newBoard[i] = b[i]
	}
	return newBoard
}

// GetCurrentPlayer determina qué jugador debe mover basándose en la cantidad de fichas en el tablero.
// Se asume que si hay igual número de fichas, le toca a las negras ('B'); de lo contrario, a las blancas ('W').
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

// EvaluateBoard calcula una puntuación que refleja qué tan favorable es el estado del tablero para el jugador.
// Se suma la "fuerza" de las cadenas de fichas del jugador y se resta la del oponente, considerando también si
// las cadenas están bloqueadas o tienen extremos abiertos.
func EvaluateBoard(b Board, player rune) float64 {
	opponent := SwitchPlayer(player)
	playerScore := 0
	oppScore := 0
	directions := []struct{ dr, dc int }{
		{0, 1}, {1, 0}, {1, 1}, {1, -1},
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

// WeightedChainScore asigna un puntaje a una cadena de fichas en función de su longitud y de si sus extremos están bloqueados.
// Los puntajes son escalados para reflejar la importancia de cadenas más largas.
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

// chainInfo examina una cadena de fichas a partir de una posición dada (r,c) en la dirección (dr,dc)
// y retorna la longitud de la cadena, además de indicar si los extremos de la misma están bloqueados.
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

// FindCriticalBlocks recorre el tablero y devuelve todas las posiciones vacías
// que, al colocar la ficha del oponente, generan una cadena "crítica" (p.ej. al menos 4 fichas consecutivas)
// en alguna dirección, considerando todas las orientaciones (horizontal, vertical y diagonales).
func FindCriticalBlocks(b Board, opponent rune) []Position {
	var crit []Position
	directions := []struct{ dr, dc int }{
		{0, 1}, {1, 0}, {1, 1}, {1, -1},
	}
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] != '\x00' {
				continue
			}
			for _, d := range directions {
				count := 1 // contando la ficha hipotética que se colocaría
				// Contar hacia adelante
				fr, fc := r, c
				for {
					fr += d.dr
					fc += d.dc
					if fr < 0 || fr >= BoardSize || fc < 0 || fc >= BoardSize {
						break
					}
					if b[fr][fc] == opponent {
						count++
					} else {
						break
					}
				}
				// Contar hacia atrás
				br, bc := r, c
				for {
					br -= d.dr
					bc -= d.dc
					if br < 0 || br >= BoardSize || bc < 0 || bc >= BoardSize {
						break
					}
					if b[br][bc] == opponent {
						count++
					} else {
						break
					}
				}
				// Si la cuenta es mayor o igual a 4, se considera crítica.
				if count >= 4 {
					crit = append(crit, Position{r, c})
					break
				}
			}
		}
	}
	return crit
}

// FindBestComplementForCritical busca, dado el tablero b, la posición crítica 'crit' y el jugador actual,
// la segunda posición vacía legal que, al ser combinada con 'crit', genere el mejor estado para bloquear al oponente.
// Se evalúa cada candidato simulando el movimiento (dos fichas) y se escoge el que maximiza EvaluateBoard(b, currentPlayer).
func FindBestComplementForCritical(b Board, currentPlayer rune, crit Position) Position {
	bestEval := -1e9 // Valor muy bajo
	var bestPos Position
	// Iterar sobre todas las posiciones vacías
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			cand := Position{r, c}
			if cand == crit {
				continue
			}
			// Verificar que la posición esté vacía y que la jugada (crit, cand) sea válida.
			if b[r][c] != '\x00' || !IsValidMove(b, crit, cand) {
				continue
			}
			simBoard := CloneBoard(b)
			var move Move = Move{crit, cand}
			ApplyMove(&simBoard, move, currentPlayer)
			eval := EvaluateBoard(simBoard, currentPlayer)
			if eval > bestEval {
				bestEval = eval
				bestPos = cand
			}
		}
	}
	return bestPos
}

// baseSmartMoves genera una lista de movimientos básicos tomando posiciones prioritarias del tablero.
// Se utiliza la función GetPriorityPositions para obtener celdas de interés y luego se forman todos los pares posibles,
// hasta un máximo de 100 movimientos. Este conjunto sirve para limitar el espacio de búsqueda a jugadas prometedoras.
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

// GenerateSmartMoves combina jugadas ganadoras inmediatas (si se detectan) con movimientos básicos
// generados por baseSmartMoves, y retorna un conjunto de movimientos potencialmente prometedores, limitado a 150.
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

// FindWinningMove examina los movimientos básicos generados por baseSmartMoves y retorna aquel que,
// al aplicarlo, resulte en una victoria inmediata para el jugador. Se descartan movimientos cuyo estado
// evaluado no supere un umbral establecido (en este caos, 2000 puntos), lo que ayuda a filtrar jugadas poco prometedoras.
func FindWinningMove(b Board, player rune) *Move {
	moves := baseSmartMoves(b)
	// Define un umbral mínimo; movimientos que produzcan un estado con evaluación inferior se descartan.
	threshold := 2000.0
	for i := 0; i < len(moves); i++ {
		testBoard := CloneBoard(b)
		ApplyMove(&testBoard, moves[i], player)
		// Si la evaluación es baja, omite este movimiento.
		if EvaluateBoard(testBoard, player) < threshold {
			continue
		}
		if CheckWin(testBoard, player) {
			return &moves[i]
		}
	}
	return nil
}

// GetPriorityPositions retorna posiciones del tablero que se consideran de mayor interés para formar jugadas.
// Si el tablero está vacío, se selecciona una región central (por ejemplo, un área 5x5). En otros casos, se recogen
// las posiciones vacías cercanas a las fichas existentes.
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

// IsBoardEmpty determina si el tablero está completamente vacío.
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

// mapToSlice convierte un mapa de posiciones en un slice para facilitar su manejo.
func mapToSlice(m map[Position]bool) []Position {
	result := make([]Position, 0, len(m))
	for pos := range m {
		result = append(result, pos)
	}
	return result
}

// GetWinner determina el ganador del juego basándose en el estado actual del tablero.
// Retorna 'B' si ganan las negras, 'W' si ganan las blancas o ' ' si no hay ganador.
func GetWinner(board Board) rune {
	if CheckWin(board, 'B') {
		return 'B'
	}
	if CheckWin(board, 'W') {
		return 'W'
	}
	return ' '
}

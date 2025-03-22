package mcts

import (
	"connect6/board"
	"math"
	"math/rand"
	"time"
)

// MCTS contiene los parámetros de la búsqueda.
type MCTS struct {
	Iterations  int     // Número máximo de simulaciones
	Exploration float64 // Factor de exploración (por ejemplo, sqrt(2))
	MaxDepth    int     // Profundidad máxima de simulación (rollout)
	TimeLimit   int     // Tiempo límite en segundos para la búsqueda (se pasa como parámetro)
}

// Node representa un nodo en el árbol MCTS.
type Node struct {
	board        board.Board  // Estado del tablero en este nodo
	parent       *Node        // Nodo padre
	children     []*Node      // Hijos expandidos
	visits       int          // Número de visitas
	wins         float64      // Número de victorias acumuladas
	untriedMoves []board.Move // Movimientos aún no explorados desde este nodo
	move         board.Move   // Movimiento que llevó a este nodo (si tiene padre)
	player       rune         // Jugador que realizó el movimiento para llegar a este estado
}

// NewNode crea un nuevo nodo a partir de un estado dado.
func NewNode(b board.Board, move board.Move, parent *Node, player rune) *Node {
	return &Node{
		board:        b,
		move:         move,
		parent:       parent,
		player:       player,
		children:     []*Node{},
		visits:       0,
		wins:         0,
		untriedMoves: []board.Move{},
	}
}

// Search inicia la búsqueda MCTS desde el estado actual (rootBoard) para el jugador que tiene el turno (currentPlayer).
// Se establece un límite de tiempo y se exploran simulaciones hasta alcanzarlo o superar el número máximo de iteraciones.
// Antes de iniciar las simulaciones, se revisan jugadas que permitan ganar inmediatamente o bloquear amenazas críticas.
func (m *MCTS) Search(rootBoard board.Board, currentPlayer rune) board.Move {
	deadline := time.Now().Add(time.Duration(m.TimeLimit) * time.Second)
	root := NewNode(rootBoard, board.Move{}, nil, currentPlayer)
	root.untriedMoves = generateLegalMoves(root.board, currentPlayer)

	// 1) Revisar si currentPlayer puede ganar de inmediato.
	winningPositions := board.FindCriticalBlocks(root.board, currentPlayer)
	if len(winningPositions) > 0 {
		var move board.Move
		if len(winningPositions) >= 2 {
			// Si hay dos o más posiciones críticas, usar las dos mejores.
			move = board.Move{winningPositions[0], winningPositions[1]}
		} else {
			// Si solo hay una, buscamos el mejor complemento entre las posiciones vacías.
			bestComplement := board.FindBestComplementForCritical(root.board, currentPlayer, winningPositions[0])
			move = board.Move{winningPositions[0], bestComplement}
		}
		return move
	}

	// Detectar amenazas críticas de forma eficiente.
	// Se buscan posiciones críticas para bloquear al oponente.
	opponent := board.SwitchPlayer(currentPlayer)
	criticalPositions := board.FindCriticalBlocks(root.board, opponent)
	if len(criticalPositions) > 0 {
		var move board.Move
		if len(criticalPositions) >= 2 {
			// Si hay dos o más posiciones críticas, usar las dos mejores.
			move = board.Move{criticalPositions[0], criticalPositions[1]}
		} else {
			// Si sólo hay una, se busca el mejor complemento entre las posiciones vacías.
			bestComplement := board.FindBestComplementForCritical(root.board, currentPlayer, criticalPositions[0])
			move = board.Move{criticalPositions[0], bestComplement}
		}
		return move
	}

	// Si no se detecta amenaza crítica, se ejecuta MCTS.
	for i := 0; i < m.Iterations; i++ {
		if time.Now().After(deadline) {
			break
		}
		node := treePolicy(root, currentPlayer)
		result := defaultPolicy(node.board, board.SwitchPlayer(node.player), m, currentPlayer)
		backup(node, result)
	}
	bestChild := selectBestChild(root, 0)
	return bestChild.move
}

// treePolicy recorre el árbol de búsqueda hasta alcanzar un nodo que no esté completamente expandido
// o se encuentre en un estado terminal, de acuerdo con la función CheckWin.
func treePolicy(node *Node, currentPlayer rune) *Node {
	for !board.CheckWin(node.board, board.SwitchPlayer(currentPlayer)) &&
		!board.CheckWin(node.board, currentPlayer) {
		if len(node.untriedMoves) > 0 {
			return expand(node, currentPlayer)
		} else if len(node.children) > 0 {
			node = selectBestChild(node, 1.414) // sqrt(2) típico
			// Se actualiza el turno del jugador después de aplicar el movimiento.
			currentPlayer = board.SwitchPlayer(node.player)
		} else {
			return node
		}
	}
	return node
}

// expand elige aleatoriamente un movimiento no explorado del nodo y crea un hijo aplicando dicho movimiento.
// Se actualiza el estado clonando el tablero actual y aplicando el movimiento correspondiente.
func expand(node *Node, currentPlayer rune) *Node {
	idx := rand.Intn(len(node.untriedMoves))
	move := node.untriedMoves[idx]
	node.untriedMoves = append(node.untriedMoves[:idx], node.untriedMoves[idx+1:]...)
	newBoard := board.CloneBoard(node.board)
	applyMoveMC(&newBoard, move, currentPlayer)
	nextPlayer := board.SwitchPlayer(currentPlayer)
	child := NewNode(newBoard, move, node, currentPlayer)
	child.untriedMoves = generateLegalMoves(child.board, nextPlayer)
	node.children = append(node.children, child)
	return child
}

// defaultPolicy simula una partida aleatoria (rollout) a partir de un estado dado, hasta una profundidad máxima.
// Se evalúa el estado terminal o, de no alcanzarlo, se utiliza EvaluateBoard para determinar la calidad del estado.
func defaultPolicy(b board.Board, simPlayer rune, m *MCTS, rootPlayer rune) float64 {
	simBoard := board.CloneBoard(b)
	currentSimPlayer := simPlayer
	depth := 0
	for depth < m.MaxDepth {
		if board.CheckWin(simBoard, board.SwitchPlayer(currentSimPlayer)) {
			if rootPlayer == board.SwitchPlayer(currentSimPlayer) {
				return 1.0
			}
			return 0.0
		} else if board.CheckWin(simBoard, currentSimPlayer) {
			if currentSimPlayer == rootPlayer {
				return 1.0
			}
			return 0.0
		}

		moves := generateLegalMoves(simBoard, currentSimPlayer)
		if len(moves) == 0 {
			break
		}
		move := moves[rand.Intn(len(moves))]
		applyMoveMC(&simBoard, move, currentSimPlayer)
		currentSimPlayer = board.SwitchPlayer(currentSimPlayer)
		depth++
	}
	eval := board.EvaluateBoard(simBoard, rootPlayer)
	if eval > 0 {
		return 1.0
	}
	return 0.0
}

// backup actualiza las estadísticas del nodo y de sus ancestros con el resultado de la simulación.
func backup(node *Node, result float64) {
	for node != nil {
		node.visits++
		node.wins += result
		// En juegos de suma cero se invierte el resultado al subir.
		result = 1.0 - result
		node = node.parent
	}
}

// selectBestChild selecciona el hijo del nodo que maximiza el valor UCB.
func selectBestChild(node *Node, exploration float64) *Node {
	var best *Node
	bestValue := -math.MaxFloat64
	for _, child := range node.children {
		ucb := child.wins/float64(child.visits) +
			exploration*math.Sqrt(math.Log(float64(node.visits))/float64(child.visits))
		if ucb > bestValue {
			bestValue = ucb
			best = child
		}
	}
	return best
}

// --- Funciones auxiliares para implementar las reglas de Connect6 ---

// generateLegalMoves genera los movimientos legales para el jugador en el estado b.
// Si es la primera jugada, se generan movimientos para colocar UNA sola ficha,
// utilizando como centinela la posición (-1,-1) en la segunda posición del movimiento.
func generateLegalMoves(b board.Board, player rune) []board.Move {
	if board.IsBoardEmpty(b) {
		var moves []board.Move
		for i := 0; i < board.BoardSize; i++ {
			for j := 0; j < board.BoardSize; j++ {
				if b[i][j] == '\x00' {
					var m board.Move
					m[0] = board.Position{Row: i, Col: j}
					m[1] = board.Position{Row: -1, Col: -1} // Centinela para indicar jugada de 1 ficha
					moves = append(moves, m)
				}
			}
		}
		return moves
	}
	// Para jugadas normales (dos fichas), se utiliza la función existente.
	return board.GenerateSmartMoves(b)
}

// applyMoveMC aplica el movimiento 'move' sobre el tablero b para el jugador dado.
// Si move[1] es (-1,-1), se coloca solo una ficha; de lo contrario, se aplican ambas fichas.
func applyMoveMC(b *board.Board, move board.Move, player rune) {
	if move[1].Row == -1 && move[1].Col == -1 {
		(*b)[move[0].Row][move[0].Col] = player
	} else {
		board.ApplyMove(b, move, player)
	}
}

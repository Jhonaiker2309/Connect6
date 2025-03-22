package mcts

import (
	"math"
	"math/rand"
	"time"

	"connect6/board"
)

type MCTS struct {
	Iterations  int     // Límite máximo de simulaciones
	Exploration float64 // Constante de exploración
	MaxDepth    int     // Profundidad máxima de la simulación (rollout)
	TimeLimit   int     // Límite en segundos
}

// Node para el árbol de búsqueda
type Node struct {
	board        board.Board
	parent       *Node
	children     []*Node
	visits       int
	wins         float64
	untriedMoves []board.Move
	move         board.Move // movimiento que llevó a este nodo
	player       rune       // jugador que hizo el movimiento 'move' en este nodo
	movesInTurn  int        // cuántos movimientos se han hecho en el turno actual (0,1,2)
}

// NewNode crea un nodo dado un estado y jugador actual
func NewNode(b board.Board, move board.Move, parent *Node, player rune, movesInTurn int) *Node {
	return &Node{
		board:        b,
		parent:       parent,
		move:         move,
		player:       player,
		movesInTurn:  movesInTurn,
		visits:       0,
		wins:         0.0,
		untriedMoves: board.GenerateSmartMoves(b),
	}
}

// Search inicia la búsqueda MCTS
func (m *MCTS) Search(state board.Board) board.Move {
	rand.Seed(time.Now().UnixNano())

	// Determinamos quién es el jugador actual
	currentPlayer := board.GetCurrentPlayer(state)
	// Creamos la raíz
	root := NewNode(state, board.Move{}, nil, board.SwitchPlayer(currentPlayer), 0)
	// Definimos 'player' como si fuera "quién movió para llegar aquí".

	// Control de tiempo: deadline
	deadline := time.Now().Add(time.Duration(m.TimeLimit) * time.Second)

	for i := 0; i < m.Iterations; i++ {
		if time.Now().After(deadline) {
			break
		}
		// 1) Selection
		node := m.selectNode(root)
		// 2) Expansion
		expanded := m.expand(node)
		// 3) Simulation (rollout)
		result := m.rollout(expanded)
		// 4) Backpropagation
		m.backpropagate(expanded, result)
	}
	// Elegimos el hijo con el mayor número de visitas (o mayor ratio wins)
	return m.getBestMove(root)
}

// selectNode recorre el árbol hasta llegar a un nodo no completamente expandido
func (m *MCTS) selectNode(node *Node) *Node {
	current := node
	for len(current.untriedMoves) == 0 && len(current.children) > 0 {
		current = m.ucbSelect(current)
	}
	return current
}

// ucbSelect elige el hijo con mayor valor UCB
func (m *MCTS) ucbSelect(node *Node) *Node {
	var bestNode *Node
	bestValue := -math.MaxFloat64

	for _, child := range node.children {
		ucb := m.ucbValue(child, node.visits)
		if ucb > bestValue {
			bestValue = ucb
			bestNode = child
		}
	}
	return bestNode
}

// ucbValue calcula UCB = (wins/visits) + C * sqrt( ln(parentVisits)/visits )
func (m *MCTS) ucbValue(node *Node, parentVisits int) float64 {
	if node.visits == 0 {
		return math.Inf(1)
	}
	exploit := node.wins / float64(node.visits)
	explore := math.Sqrt(math.Log(float64(parentVisits)) / float64(node.visits))
	return exploit + m.Exploration*explore
}

func (m *MCTS) expand(node *Node) *Node {
	if len(node.untriedMoves) == 0 {
		return node
	}
	moveIdx := rand.Intn(len(node.untriedMoves))
	move := node.untriedMoves[moveIdx]
	node.untriedMoves = append(node.untriedMoves[:moveIdx], node.untriedMoves[moveIdx+1:]...)

	newBoard := board.CloneBoard(node.board)
	currentPlayer := node.player
	movesInTurn := node.movesInTurn + 1

	// Cambia de jugador tras 2 movimientos (Connect6)
	if movesInTurn == 2 {
		currentPlayer = board.SwitchPlayer(node.player)
		movesInTurn = 0
	}

	board.ApplyMove(&newBoard, move, currentPlayer)

	child := NewNode(newBoard, move, node, currentPlayer, movesInTurn)
	node.children = append(node.children, child)
	return child
}

// rollout ejecuta la fase de simulación hasta MaxDepth o estado terminal
func (m *MCTS) rollout(node *Node) float64 {
	state := board.CloneBoard(node.board)
	currentPlayer := node.player
	originalPlayer := currentPlayer
	movesInTurn := node.movesInTurn

	for depth := 0; depth < m.MaxDepth; depth++ {
		// Verificar si alguien ganó
		if board.CheckWin(state, 'B') {
			if originalPlayer == 'B' {
				return 1.0
			} else {
				return 0.0
			}
		} else if board.CheckWin(state, 'W') {
			if originalPlayer == 'W' {
				return 1.0
			} else {
				return 0.0
			}
		}

		moves := board.GenerateSmartMoves(state)
		if len(moves) == 0 {
			break
		}

		// Realizar 2 movimientos (connect6)
		for i := 0; i < 2; i++ {
			if len(moves) == 0 {
				break
			}

			move := m.policyMove(state, moves, currentPlayer)
			board.ApplyMove(&state, move, currentPlayer)
			movesInTurn++

			if board.CheckWin(state, 'B') || board.CheckWin(state, 'W') {
				break
			}

			if movesInTurn == 2 {
				currentPlayer = board.SwitchPlayer(currentPlayer)
				movesInTurn = 0
			}
		}
	}

	eval := board.EvaluateBoard(state, originalPlayer)
	if eval <= 0 {
		return 0.0
	}
	return 1.0
}

// policyMove: elige un movimiento durante la simulación.
// 1) Jugada ganadora
// 2) Bloqueo
// 3) EvaluateBoard
// Con pequeña aleatoriedad
func (m *MCTS) policyMove(state board.Board, moves []board.Move, currentPlayer rune) board.Move {
	// 1) Movida ganadora tuya
	if winMove := board.FindPairWinningMove(state, currentPlayer); winMove != nil {
		return *winMove
	}
	// 2) Bloqueo movida ganadora rival
	opponent := board.SwitchPlayer(currentPlayer)
	if blockMove := board.FindPairWinningMove(state, opponent); blockMove != nil {
		return *blockMove
	}

	// 3) Bloqueo “4 en línea con 2 huecos” del rival
	//    => Escanear jugadas que disminuyan la evaluación del rival
	var bestBlockMove *board.Move
	bestBlockEval := math.Inf(1) // mientras más bajo, mejor para nosotros

	for _, mv := range moves {
		tmp := board.CloneBoard(state)
		// Jugada nuestra
		board.ApplyMove(&tmp, mv, currentPlayer)
		// Evaluación del rival tras esto
		oppVal := board.EvaluateBoard(tmp, opponent)
		if oppVal < bestBlockEval {
			bestBlockEval = oppVal
			copyMove := mv
			bestBlockMove = &copyMove
		}
	}

	// Revisamos si bestBlockEval es muy inferior => es un buen “bloqueo”
	// Podemos poner un umbral, p.ej. si EvaluateBoard del rival normal era 20000,
	// y con nuestro bestBlockEval bajamos a 5000, es un gran bloqueo.
	// Sino, seguimos con la heurística normal.

	if bestBlockMove != nil {
		// Checa la evaluación del rival en la posición actual
		currentRivalVal := board.EvaluateBoard(state, opponent)
		// si la diferencia es grande, bloquea
		if currentRivalVal-bestBlockEval > 10000 {
			// => hay un gran cambio => haremos ese blocking
			return *bestBlockMove
		}
	}

	// 4) Heurística “positiva” => elegimos la que me da mejor EvaluateBoard
	bestScore := -math.MaxFloat64
	bestMove := moves[rand.Intn(len(moves))]

	for _, mv := range moves {
		tmp := board.CloneBoard(state)
		board.ApplyMove(&tmp, mv, currentPlayer)
		sc := board.EvaluateBoard(tmp, currentPlayer)
		if sc > bestScore {
			bestScore = sc
			bestMove = mv
		}
	}

	return bestMove
}

// backpropagate recorre hacia arriba y ajusta visits/wins
func (m *MCTS) backpropagate(node *Node, result float64) {
	current := node
	for current != nil {
		current.visits++
		current.wins += result

		current = current.parent
	}
}

// getBestMove elige el movimiento en el hijo con el mayor número de visitas
func (m *MCTS) getBestMove(root *Node) board.Move {
	// 1) Buscar jugadas ganadoras en profundidad 1
	for _, child := range root.children {
		if child.movesInTurn == 0 {
			// si completó 2 movidas
			if board.CheckWin(child.board, root.player) {
				return child.move
			}
		}
	}

	// 2) Elegir el hijo más visitado
	var bestChild *Node
	bestVisits := -1
	for _, child := range root.children {
		if child.visits > bestVisits {
			bestVisits = child.visits
			bestChild = child
		}
	}

	if bestChild == nil {
		moves := board.GenerateSmartMoves(root.board)
		if len(moves) == 0 {
			return board.Move{}
		}
		return moves[0]
	}
	return bestChild.move
}

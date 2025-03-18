package mcts

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"connect6/board"
)

// Estructura para guardar estadísticas de transposición.
type TTEntry struct {
	visits int
	wins   float64
}

type TranspositionTable struct {
	sync.RWMutex
	table map[string]*TTEntry
}

func NewTranspositionTable() *TranspositionTable {
	return &TranspositionTable{
		table: make(map[string]*TTEntry),
	}
}

func (tt *TranspositionTable) Get(state board.Board) *TTEntry {
	tt.RLock()
	defer tt.RUnlock()
	key := board.BoardHash(state) // Necesitarás la función de hash
	return tt.table[key]
}

func (tt *TranspositionTable) Update(state board.Board, visits int, wins float64) {
	tt.Lock()
	defer tt.Unlock()
	key := board.BoardHash(state)
	entry, ok := tt.table[key]
	if !ok {
		tt.table[key] = &TTEntry{visits: visits, wins: wins}
	} else {
		entry.visits += visits
		entry.wins += wins
	}
}

// MCTS contiene parámetros para el algoritmo
type MCTS struct {
	Iterations       int     // Límite máximo de simulaciones
	Exploration      float64 // Constante de exploración
	MaxDepth         int     // Profundidad máxima de la simulación (rollout)
	UseTransposition bool
	TT               *TranspositionTable
	TimeLimit        int // Límite en segundos
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
}

// NewNode crea un nodo dado un estado y jugador actual
func NewNode(b board.Board, move board.Move, parent *Node, player rune) *Node {
	return &Node{
		board:        b,
		parent:       parent,
		move:         move,
		player:       player,
		visits:       0,
		wins:         0.0,
		untriedMoves: board.GenerateSmartMoves(b),
	}
}

// Search inicia la búsqueda MCTS
func (m *MCTS) Search(state board.Board) board.Move {
	rand.Seed(time.Now().UnixNano())

	// Si vamos a usar tabla de transposición, inicializamos si no existe
	if m.TT == nil && m.UseTransposition {
		m.TT = NewTranspositionTable()
	}

	// Determinamos quién es el jugador actual
	currentPlayer := board.GetCurrentPlayer(state)

	// Creamos la raíz
	root := NewNode(state, board.Move{}, nil, board.SwitchPlayer(currentPlayer))
	// Definimos 'player' como si fuera "quién movió para llegar aquí".

	// === Control de tiempo: calculamos deadline ===
	deadline := time.Now().Add(time.Duration(m.TimeLimit) * time.Second)

	for i := 0; i < m.Iterations; i++ {
		// Verificamos si se acabó el tiempo
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

	// Elegimos el hijo con el mayor número de visitas (o mayor ratio de wins)
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

// expand toma uno de los movimientos no intentados y crea un hijo
func (m *MCTS) expand(node *Node) *Node {
	if len(node.untriedMoves) == 0 {
		return node
	}
	moveIdx := rand.Intn(len(node.untriedMoves))
	move := node.untriedMoves[moveIdx]
	node.untriedMoves = append(node.untriedMoves[:moveIdx], node.untriedMoves[moveIdx+1:]...)

	newBoard := board.CloneBoard(node.board)
	currentPlayer := board.SwitchPlayer(node.player)
	board.ApplyMove(&newBoard, move, currentPlayer)

	child := NewNode(newBoard, move, node, currentPlayer)
	node.children = append(node.children, child)
	return child
}

// rollout - fase de simulación - jugamos hasta maxDepth o posición terminal
func (m *MCTS) rollout(node *Node) float64 {
	state := board.CloneBoard(node.board)
	currentPlayer := board.GetCurrentPlayer(state)
	originalPlayer := currentPlayer

	for depth := 0; depth < m.MaxDepth; depth++ {
		// Verificamos si alguien ganó
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

		// Elige un movimiento según 'policyMove'
		move := m.policyMove(state, moves, currentPlayer)
		board.ApplyMove(&state, move, currentPlayer)
		currentPlayer = board.SwitchPlayer(currentPlayer)
	}

	eval := board.EvaluateBoard(state, originalPlayer)
	if eval <= 0 {
		return 0.0
	}
	return 1.0
}

// policyMove elige un movimiento durante la simulación de forma informada
func (m *MCTS) policyMove(state board.Board, moves []board.Move, currentPlayer rune) board.Move {
	// 1) Jugada ganadora
	if winMove := board.FindWinningMove(state, currentPlayer); winMove != nil {
		return *winMove
	}
	// 2) Bloquear
	if blockMove := board.FindWinningMove(state, board.SwitchPlayer(currentPlayer)); blockMove != nil {
		return *blockMove
	}

	// 3) Heurística simple
	bestScore := -999999.0
	bestMove := moves[rand.Intn(len(moves))] // fallback

	for _, mv := range moves {
		tmp := board.CloneBoard(state)
		board.ApplyMove(&tmp, mv, currentPlayer)
		sc := board.EvaluateBoard(tmp, currentPlayer)
		if sc > bestScore {
			bestScore = sc
			bestMove = mv
		}
	}

	// para no ser determinista: 50% la mejor, 50% aleatorio
	if rand.Float64() < 0.5 {
		return bestMove
	}
	return moves[rand.Intn(len(moves))]
}

// backpropagate recorre hacia arriba y ajusta visits/wins
func (m *MCTS) backpropagate(node *Node, result float64) {
	current := node
	for current != nil {
		current.visits++
		current.wins += result

		if m.UseTransposition && m.TT != nil {
			m.TT.Update(current.board, 1, result)
		}
		current = current.parent
	}
}

// getBestMove elige el movimiento en el hijo con el mayor número de visitas
func (m *MCTS) getBestMove(root *Node) board.Move {
	var bestChild *Node
	bestVisits := -1

	for _, child := range root.children {
		if child.visits > bestVisits {
			bestVisits = child.visits
			bestChild = child
		}
	}
	if bestChild == nil {
		// si no hay hijos, generamos alguno
		moves := board.GenerateSmartMoves(root.board)
		if len(moves) == 0 {
			return board.Move{}
		}
		return moves[0]
	}
	return bestChild.move
}

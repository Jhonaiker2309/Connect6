package mcts

import (
	"math"
	"math/rand"
	"time"
	"connect6/board"
)

const (
	maxDepth = 20
)

type Node struct {
	board        board.Board
	parent       *Node
	children     []*Node
	visits       int
	wins         float64
	untriedMoves []board.Move
	move         board.Move
}

type MCTS struct {
	Iterations   int
	Exploration  float64
}

func (m *MCTS) Search(state board.Board) board.Move {
	root := &Node{
		board:        state,
		untriedMoves: board.GenerateSmartMoves(state),
	}

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < m.Iterations; i++ {
		node := m.selectNode(root)
		node = m.expand(node)
		result := m.simulate(node)
		m.backpropagate(node, result)
	}

	return m.getBestMove(root)
}

func (m *MCTS) selectNode(node *Node) *Node {
	current := node
	for len(current.untriedMoves) == 0 && len(current.children) > 0 {
		current = m.ucbSelect(current)
	}
	return current
}

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

func (m *MCTS) ucbValue(node *Node, parentVisits int) float64 {
	if node.visits == 0 {
		return math.Inf(1)
	}
	return (node.wins / float64(node.visits)) +
		m.Exploration*math.Sqrt(math.Log(float64(parentVisits))/float64(node.visits))
}

func (m *MCTS) expand(node *Node) *Node {
	if len(node.untriedMoves) == 0 {
		return node
	}

	moveIdx := rand.Intn(len(node.untriedMoves))
	move := node.untriedMoves[moveIdx]
	node.untriedMoves = append(node.untriedMoves[:moveIdx], node.untriedMoves[moveIdx+1:]...)

	newBoard := board.CloneBoard(node.board)
	board.ApplyMove(&newBoard, move, board.GetCurrentPlayer(node.board))

	child := &Node{
		board:        newBoard,
		parent:       node,
		untriedMoves: board.GenerateSmartMoves(newBoard),
		move:         move,
	}
	node.children = append(node.children, child)
	return child
}

func (m *MCTS) simulate(node *Node) float64 {
	state := board.CloneBoard(node.board)
	currentPlayer := board.GetCurrentPlayer(state)
	originalPlayer := currentPlayer

	for depth := 0; depth < maxDepth; depth++ {
		if board.CheckWin(state, 'B') || board.CheckWin(state, 'W') {
			break
		}

		moves := board.GenerateSmartMoves(state)
		if len(moves) == 0 {
			break
		}

		var move board.Move
		if winMove := board.FindWinningMove(state, currentPlayer); winMove != nil {
			move = *winMove
		} else if blockMove := board.FindWinningMove(state, board.SwitchPlayer(currentPlayer)); blockMove != nil {
			move = *blockMove
		} else {
			move = moves[rand.Intn(len(moves))]
		}

		board.ApplyMove(&state, move, currentPlayer)
		currentPlayer = board.SwitchPlayer(currentPlayer)
	}

	return board.EvaluateBoard(state, originalPlayer)
}

func (m *MCTS) backpropagate(node *Node, result float64) {
	for n := node; n != nil; n = n.parent {
		n.visits++
		n.wins += result
	}
}

func (m *MCTS) getBestMove(root *Node) board.Move {
	var bestMove board.Move
	maxVisits := -1

	for _, child := range root.children {
		if child.visits > maxVisits {
			maxVisits = child.visits
			bestMove = child.move
		}
	}
	return bestMove
}
package game

import (
	"connect6/board"
	"connect6/mcts"
	"connect6/ui"
	"fmt"
	"math/rand"
	"time"
)

// Game representa la instancia principal del juego Connect6
// Contiene el estado del tablero, el motor de IA y el jugador actual
type Game struct {
	board         board.Board
	mcts          *mcts.MCTS
	currentPlayer rune
	tpj           int
}

// NewGame crea e inicializa una nueva instancia del juego
// Retorna:
//   - Puntero a Game configurado y listo para iniciar
func NewGame(fichas string, tiempo int) *Game {
	rand.Seed(time.Now().UnixNano())

	// Decide quién inicia, según '-fichas='
	var initialPlayer rune
	if fichas == "blancas" {
		initialPlayer = 'W'
	} else {
		// default: negras
		initialPlayer = 'B'
	}

	return &Game{
		mcts: &mcts.MCTS{
			MaxDepth:         30,
			Iterations:       10000,
			Exploration:      1.414, // sqrt(2)
			UseTransposition: true,
			TimeLimit:        tiempo,
		},
		currentPlayer: initialPlayer,
		tpj:           tiempo,
	}
}

// Run ejecuta el bucle principal del juego
// Flujo:
//  1. Muestra el tablero
//  2. Verifica victoria
//  3. Alterna turnos entre jugador y IA
//  4. Finaliza cuando hay un ganador
func (g *Game) Run() {
	for {
		ui.PrintBoard(g.board)

		if board.CheckWin(g.board, 'B') || board.CheckWin(g.board, 'W') {
			break
		}

		if g.currentPlayer == 'B' {
			g.botTurn()
		} else {
			g.playerTurn()
		}

		g.currentPlayer = board.SwitchPlayer(g.currentPlayer)
	}
	g.showFinalResult()
}

// botTurn maneja el turno de la IA
// Pasos:
//  1. Ejecuta la búsqueda MCTS para encontrar el mejor movimiento
//  2. Aplica el movimiento al tablero
func (g *Game) botTurn() {
	fmt.Println("Turno del Bot (Negras)...")
	bestMove := g.mcts.Search(g.board) // Obtiene mejor movimiento de la IA
	board.ApplyMove(&g.board, bestMove, 'B')
}

// playerTurn maneja el turno del jugador humano
// Pasos:
//  1. Solicita entrada al jugador
//  2. Valida y aplica el movimiento
func (g *Game) playerTurn() {
	fmt.Println("Tu turno (Blancas)")
	move := ui.GetPlayerMove(g.board) // Obtiene movimiento del jugador
	board.ApplyMove(&g.board, move, 'W')
}

// showFinalResult muestra el resultado final del juego
// - Imprime el tablero final
// - Muestra mensaje de victoria/empate
func (g *Game) showFinalResult() {
	ui.PrintBoard(g.board)
	ui.ShowResult(board.GetWinner(g.board))
}

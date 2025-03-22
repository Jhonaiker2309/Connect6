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
	humanPiece    rune
}

// NewGame crea e inicializa una nueva instancia del juego
// Retorna:
//   - Puntero a Game configurado y listo para iniciar
func NewGame(fichas string, tiempo int) *Game {
	rand.Seed(time.Now().UnixNano())

    // Negro siempre inicia.
    initialPlayer := 'B'
    var humanPiece rune
    if fichas == "negras" {
        humanPiece = 'B'
    } else {
        humanPiece = 'W'
    }

	return &Game{
		mcts: &mcts.MCTS{
			MaxDepth:    30,
			Iterations:  100000,
			Exploration: 1.414, // sqrt(2)
			TimeLimit:   tiempo,
		},
		currentPlayer: initialPlayer,
		tpj:           tiempo,
		humanPiece:    humanPiece,
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

        if g.currentPlayer == g.humanPiece {
            g.playerTurn()
        } else {
            g.botTurn()
        }

		g.currentPlayer = board.SwitchPlayer(g.currentPlayer)
	}
	g.showFinalResult()
}

// botTurn maneja el turno de la IA
// Pasos:
//  1. Ejecuta la búsqueda MCTS para encontrar el mejor movimiento
//  2. Aplica el movimiento al tablero
// botTurn maneja el turno de la IA
// El Bot juega con la pieza opuesta a la del jugador humano.
func (g *Game) botTurn() {
    fmt.Println("Turno del Bot...")
    // El Bot siempre juega con la pieza opuesta
    var botPiece rune
    if g.humanPiece == 'B' {
        botPiece = 'W'
    } else {
        botPiece = 'B'
    }
    bestMove := g.mcts.Search(g.board, g.currentPlayer)
    board.ApplyMove(&g.board, bestMove, botPiece)
}

// playerTurn maneja el turno del jugador humano
// Pasos:
//  1. Solicita entrada al jugador
//  2. Valida y aplica el movimiento
func (g *Game) playerTurn() {
    fmt.Println("Tu turno...")
    move := ui.GetPlayerMove(g.board)
    board.ApplyMove(&g.board, move, g.humanPiece)
}

// showFinalResult muestra el resultado final del juego
// - Imprime el tablero final
// - Muestra mensaje de victoria/empate
func (g *Game) showFinalResult() {
	ui.PrintBoard(g.board)
	ui.ShowResult(board.GetWinner(g.board))
}

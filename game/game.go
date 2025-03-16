package game

import (
    "connect6/board"
    "connect6/mcts"
    "connect6/ui"
    "math/rand"
    "time"
	"fmt"
)

// Game representa la instancia principal del juego Connect6
// Contiene el estado del tablero, el motor de IA y el jugador actual
type Game struct {
    board         board.Board
    mcts          *mcts.MCTS
    currentPlayer rune
}

// NewGame crea e inicializa una nueva instancia del juego
// Retorna:
//   - Puntero a Game configurado y listo para iniciar
func NewGame() *Game {
    rand.Seed(time.Now().UnixNano())
    return &Game{
        mcts: &mcts.MCTS{
            Iterations:  1000,      // Número de simulaciones MCTS por turno
            Exploration: 1.414,    // Parámetro de exploración (√2)
        },
        currentPlayer: ui.ShowGameMenu(),  // Determina quién inicia el juego
    }
}

// Run ejecuta el bucle principal del juego
// Flujo:
//   1. Muestra el tablero
//   2. Verifica victoria
//   3. Alterna turnos entre jugador y IA
//   4. Finaliza cuando hay un ganador
func (g *Game) Run() {
    for {
        ui.PrintBoard(g.board)
        
        if board.CheckWin(g.board, 'B') || board.CheckWin(g.board, 'W') {
            break
        }
        
        if g.currentPlayer == 'B' {
            g.botTurn()    // Turno de la IA (Negras)
        } else {
            g.playerTurn() // Turno del jugador (Blancas)
        }
        
        g.currentPlayer = board.SwitchPlayer(g.currentPlayer)
    }
    g.showFinalResult()
}

// botTurn maneja el turno de la IA
// Pasos:
//   1. Ejecuta la búsqueda MCTS para encontrar el mejor movimiento
//   2. Aplica el movimiento al tablero
func (g *Game) botTurn() {
    fmt.Println("Turno del Bot (Negras)...")
    bestMove := g.mcts.Search(g.board)    // Obtiene mejor movimiento de la IA
    board.ApplyMove(&g.board, bestMove, 'B')
}

// playerTurn maneja el turno del jugador humano
// Pasos:
//   1. Solicita entrada al jugador
//   2. Valida y aplica el movimiento
func (g *Game) playerTurn() {
    fmt.Println("Tu turno (Blancas)")
    move := ui.GetPlayerMove(g.board)     // Obtiene movimiento del jugador
    board.ApplyMove(&g.board, move, 'W')
}

// showFinalResult muestra el resultado final del juego
// - Imprime el tablero final
// - Muestra mensaje de victoria/empate
func (g *Game) showFinalResult() {
    ui.PrintBoard(g.board)
    ui.ShowResult(board.GetWinner(g.board))
}
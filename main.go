package main

import (
	"connect6/game"
	"flag"
	"fmt"
)

// Variables globales, o inline en main()
var (
	fichasFlag string
	tpjFlag    int
)

func init() {
	// Define tus banderas y valores por defecto:
	flag.StringVar(&fichasFlag, "fichas", "negras", "Indica si juegas con blancas o negras")
	flag.IntVar(&tpjFlag, "tpj", 4, "Tiempo máximo (segundos) para la jugada de la IA")
}

func main() {
	// Parseamos los flags:
	flag.Parse()

	// Muestra qué se parseó (opcional)
	fmt.Println("Opción fichas:", fichasFlag)
	fmt.Println("Opción tpj:", tpjFlag)

	// Crea el juego y pásale esos parámetros
	// Para que tu agente sepa si inicia con negras o blancas y/o
	// para que la IA use 'tpjFlag' segundos por turno
	g := game.NewGame(fichasFlag, tpjFlag)
	g.Run()
}

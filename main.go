package main

import (
	"fmt"

	"github.com/brighamskarda/chess"
)

func main() {
	fmt.Println("hello world")
	chessPos, _ := chess.ParseFen(chess.DefaultFen)
	fmt.Printf("%s\n", chessPos)
}

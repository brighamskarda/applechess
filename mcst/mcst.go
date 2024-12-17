package mcst

import "github.com/brighamskarda/chess"

type Mcst struct {
	Duration int
}

func (m Mcst) GetMove(chess.Position) chess.Move {
	return chess.Move{}
}

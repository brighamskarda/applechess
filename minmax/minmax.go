package minmax

import (
	"math"

	"github.com/brighamskarda/chess"
)

type Minmax struct {
	Depth int
}

func (mm Minmax) GetMove(p chess.Position) chess.Move {
	move, _ := search(p, mm.Depth)
	return move
}

func search(p chess.Position, depth int) (chess.Move, float64) {
	if p.Turn == chess.White {
		return max(&p, depth)
	}
	if p.Turn == chess.Black {
		return min(&p, depth)
	}
	return chess.Move{}, 0
}

func min(p *chess.Position, depth int) (chess.Move, float64) {
	if depth == 0 {
		lowestScore := math.MaxFloat64
		bestMove := chess.Move{}
		for _, move := range chess.GenerateLegalMoves(p) {
			newPos := *p
			newPos.Move(move)
			score := evaluate(&newPos)
			if score < lowestScore {
				lowestScore = score
				bestMove = move
			}
		}
		return bestMove, lowestScore
	}
	lowestScore := math.MaxFloat64
	bestMove := chess.Move{}
	for _, move := range chess.GenerateLegalMoves(p) {
		newPos := *p
		newPos.Move(move)
		if chess.IsCheckMate(&newPos) {
			return move, -math.MaxFloat64
		} else if chess.IsStaleMate(&newPos) && lowestScore > 0 {
			lowestScore = 0
			bestMove = move
		} else {
			_, score := search(newPos, depth-1)
			if score < lowestScore {
				lowestScore = score
				bestMove = move
			}
		}
	}
	return bestMove, lowestScore
}

func max(p *chess.Position, depth int) (chess.Move, float64) {
	if depth == 0 {
		highestScore := -math.MaxFloat64
		bestMove := chess.Move{}
		for _, move := range chess.GenerateLegalMoves(p) {
			newPos := *p
			newPos.Move(move)
			score := evaluate(&newPos)
			if score > highestScore {
				highestScore = score
				bestMove = move
			}
		}
		return bestMove, highestScore
	}
	highestScore := -math.MaxFloat64
	bestMove := chess.Move{}
	for _, move := range chess.GenerateLegalMoves(p) {
		newPos := *p
		newPos.Move(move)
		if chess.IsCheckMate(&newPos) {
			return move, math.MaxFloat64
		} else if chess.IsStaleMate(&newPos) && highestScore < 0 {
			highestScore = 0
			bestMove = move
		} else {
			_, score := search(newPos, depth-1)
			if score > highestScore {
				highestScore = score
				bestMove = move
			}
		}
	}
	return bestMove, highestScore
}

func evaluate(p *chess.Position) float64 {
	total := sumMaterial(p)
	total += float64(numPseudoLegalChecks(p)) * 0.2
	return total
}

func sumMaterial(p *chess.Position) float64 {
	totalValue := 0.0
	for _, piece := range p.Board {
		totalValue += getPieceValue(piece)
	}
	return totalValue
}

func getPieceValue(p chess.Piece) float64 {
	const pawn = 1
	const rook = 5
	const knight = 2.9
	const bishop = 3
	const queen = 8
	const king = 10000

	var val float64
	switch p.Type {
	case chess.Pawn:
		val = pawn
	case chess.Rook:
		val = rook
	case chess.Knight:
		val = knight
	case chess.Bishop:
		val = bishop
	case chess.Queen:
		val = queen
	case chess.King:
		val = king
	default:
		val = 0
	}
	if p.Color == chess.White {
		return val
	} else {
		return -val
	}
}

func numPseudoLegalChecks(p *chess.Position) int {
	origTurn := p.Turn

	total := 0
	p.Turn = chess.White
	blackKing := findKing(p, chess.Black)
	pLegalMoves := chess.GeneratePseudoLegalMoves(p)
	for _, move := range pLegalMoves {
		if move.ToSquare == blackKing {
			total++
		}
	}

	p.Turn = chess.Black
	whiteKing := findKing(p, chess.White)
	pLegalMoves = chess.GeneratePseudoLegalMoves(p)
	for _, move := range pLegalMoves {
		if move.ToSquare == whiteKing {
			total--
		}
	}

	p.Turn = origTurn
	return total
}

func findKing(p *chess.Position, c chess.Color) chess.Square {
	for _, square := range chess.AllSquares {
		piece := p.PieceAt(square)
		if piece.Type == chess.King && piece.Color == c {
			return square
		}
	}
	return chess.NoSquare
}

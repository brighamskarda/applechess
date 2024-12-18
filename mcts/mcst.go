package mcts

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"time"

	"github.com/brighamskarda/chess"
)

const c = math.Sqrt2
const iterationsBetweenTimeChecks = 100
const randomRolloutLength = 20

// Mcts (Monte Carlo Tree Search) agent for chess
type Mcts struct {
	Duration int   // Seconds to perform search
	n        int64 // Initialize to 0.
}

type node struct {
	w        float64
	n        int64
	mov      chess.Move // The move that resulted in pos
	pos      *chess.Position
	children []*node
}

func makeParentNode(p chess.Position) *node {
	legalMoves := chess.GenerateLegalMoves(&p)
	parentNode := &node{
		w:        0,
		n:        0,
		mov:      chess.Move{},
		pos:      &p,
		children: make([]*node, 0, len(legalMoves)),
	}

	for _, move := range legalMoves {
		newPos := p
		p.Move(move)
		parentNode.children = append(parentNode.children, &node{
			w:        0,
			n:        0,
			mov:      move,
			pos:      &newPos,
			children: make([]*node, 0),
		})
	}
	return parentNode
}

func (mcts Mcts) GetMove(p chess.Position) chess.Move {
	defer mcts.resetN()
	parentNode := makeParentNode(p)

	returnChannels := make([]chan struct{}, 0, len(parentNode.children))
	for i, child := range parentNode.children {
		returnChannels = append(returnChannels, make(chan struct{}))
		go concurrentIterate(Mcts{Duration: mcts.Duration}, child, p.Turn, returnChannels[i])
	}

	for _, ch := range returnChannels {
		<-ch
	}

	var totalIterations int64
	for _, child := range parentNode.children {
		totalIterations += child.n
	}

	slog.Info("Performed " + fmt.Sprint(totalIterations) + " iterations of mcts")
	return bestMove(parentNode)
}

func concurrentIterate(mcts Mcts, n *node, agentColor chess.Color, signalDone chan struct{}) {
	startTime := time.Now()
	for time.Now().Sub(startTime).Milliseconds() < int64(mcts.Duration)*1000 {
		for i := 0; i < iterationsBetweenTimeChecks; i++ {
			n.w += mcts.iterate(n, agentColor)
			n.n++
			mcts.n++
		}
	}

	signalDone <- struct{}{}
}

// iterate is recursive, and returns 1 for a win, 0.5 for stalemate and 0 for a loss. Used for back propagation
func (mcts Mcts) iterate(n *node, agentColor chess.Color) float64 {
	if chess.IsCheckMate(n.pos) && n.pos.Turn != agentColor {
		n.n++
		n.w++
		return 1
	}
	if chess.IsStaleMate(n.pos) || chess.IsCheckMate(n.pos) {
		n.n++
		return 0
	}
	if len(n.children) == 0 {
		fillInChildren(n)
	}

	selectedNode := mcts.selectNode(n)
	var result float64
	if selectedNode.n == 0 {
		result = randomRollout(*n.pos, agentColor)
	} else {
		result = mcts.iterate(selectedNode, agentColor)
	}

	selectedNode.n++
	selectedNode.w += result
	return result
}

func (mcts Mcts) selectNode(n *node) *node {
	for _, child := range n.children {
		if child.n == 0 {
			return child
		}
	}
	maxUCB := -math.MaxFloat64
	bestChild := n.children[0]
	for _, child := range n.children {
		ucb := mcts.calcUCB(child)
		if ucb > maxUCB {
			maxUCB = ucb
			bestChild = child
		}
	}

	return bestChild
}

// randomRollout returns 1 if the agent wins, 0.5 for draw, 0 otherwise
func randomRollout(p chess.Position, agentColor chess.Color) float64 {
	for i := 0; i < randomRolloutLength; i++ {
		if chess.IsCheckMate(&p) && p.Turn != agentColor {
			return 1
		}
		if chess.IsCheckMate(&p) && p.Turn == agentColor {
			return 0
		}
		legalMoves := chess.GenerateLegalMoves(&p)
		if len(legalMoves) == 0 {
			return 0.5
		}
		move := legalMoves[rand.IntN(len(legalMoves))]
		p.Move(move)
	}
	return determineReward(&p, agentColor)
}

func determineReward(p *chess.Position, agentColor chess.Color) float64 {
	const limitForWin = 8
	positionValue := getPositionValue(p)
	switch agentColor {
	case chess.White:
		if positionValue > limitForWin {
			return 1
		}
		if positionValue < -limitForWin {
			return 0
		}
	case chess.Black:
		if positionValue > limitForWin {
			return 0
		}
		if positionValue < -limitForWin {
			return 1
		}
	}
	return 0.5
}

func getPositionValue(p *chess.Position) float64 {
	const pawn = 1
	const rook = 5
	const knight = 2.9
	const bishop = 3
	const queen = 8
	const king = 10000

	totalValue := 0.0
	for _, piece := range p.Board {
		var val float64
		switch piece.Type {
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
		if piece.Color == chess.White {
			totalValue += val
		} else if piece.Color == chess.Black {
			totalValue -= val
		}
	}
	return totalValue
}

func fillInChildren(n *node) {
	legalMoves := chess.GenerateLegalMoves(n.pos)
	for _, move := range legalMoves {
		newPos := *n.pos
		newPos.Move(move)
		newChild := &node{
			w:        0,
			n:        0,
			mov:      move,
			pos:      &newPos,
			children: make([]*node, 0),
		}
		n.children = append(n.children, newChild)
	}
}

// calcUCB uses this formula https://en.wikipedia.org/wiki/Monte_Carlo_tree_search#Exploration_and_exploitation
func (mcts Mcts) calcUCB(n *node) float64 {
	return float64(n.w)/float64(n.n) + c*math.Sqrt(math.Log(float64(mcts.n))/float64(n.n))
}

// bestMove is for selecting the best move only after all the iterations are complete.
func bestMove(n *node) chess.Move {
	bestMove := n.children[0].mov
	var bestMoveScore float64 = -math.MaxFloat64
	for _, child := range n.children {
		score := float64(child.w) / float64(child.n)
		if score > bestMoveScore {
			bestMoveScore = score
			bestMove = child.mov
		}
	}
	return bestMove
}

func (mcts Mcts) resetN() {
	mcts.n = 0
}

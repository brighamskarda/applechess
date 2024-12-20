package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/brighamskarda/applechess.git/alphabeta"
	"github.com/brighamskarda/applechess.git/mcts"
	"github.com/brighamskarda/applechess.git/minmax"
	"github.com/brighamskarda/chess"
)

func main() {
	agents := parseArgs()

	game := chess.NewGame()

	for !game.IsCheckMate() && !game.CanClaimDraw() {
		game.PrintPosition()
		move := chess.Move{}
		if game.Turn() == chess.White {
			fmt.Println("White's move")
			move = agents[0].GetMove(*game.Position())
		} else if game.Turn() == chess.Black {
			fmt.Println("Black's move")
			move = agents[1].GetMove(*game.Position())
		} else {
			slog.Error("game.Turn() is not black or white")
			os.Exit(1)
		}
		if game.Move(move) != nil {
			slog.Error("agent provided invalid move", "agent-color", game.Turn())
			os.Exit(1)
		}
		fmt.Println(move)
		fmt.Println()
	}

	if game.IsCheckMate() {
		switch game.Turn() {
		case chess.Black:
			fmt.Println("White Wins!")
		case chess.White:
			fmt.Println("Black Wins!")
		}
		os.Exit(0)
	}

	if game.CanClaimDraw() {
		fmt.Println("The Game has been automatically Drawn")
		os.Exit(1)
	}

	slog.Error("the program ended without checkmate or draw")
	os.Exit(1)
}

type ChessAgent interface {
	GetMove(chess.Position) chess.Move
}

func parseArgs() [2]ChessAgent {
	help := flag.Bool("help", false, "prints help")
	player1 := flag.String("p1", "human", "agent to play white [human|mcts|minmax|ab]")
	player2 := flag.String("p2", "human", "agent to play black [human|mcts|minmax|ab]")
	player1Option := flag.Int("o1", 2, "option for player1, for depth based agents this the depth, for time based agents this is the time in seconds")
	player2Option := flag.Int("o2", 2, "option for player2, for depth based agents this the depth, for time based agents this is the time in seconds")
	logLevel := flag.String("log", "ERROR", "logging level [ERROR|WARN|INFO|DEBUG]")

	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	switch strings.ToUpper(*logLevel) {
	case "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	case "WARN":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "INFO":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "DEBUG":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	default:
		slog.SetLogLoggerLevel(slog.LevelError)
		slog.Error("could not parse log argument", "arg", *logLevel)
	}
	agents := [2]ChessAgent{}

	switch strings.ToLower(*player1) {
	case "human":
		agents[0] = Human{}
	case "mcts":
		agents[0] = mcts.Mcts{Duration: *player1Option}
	case "minmax":
		agents[0] = minmax.Minmax{Depth: *player1Option}
	case "ab":
		agents[0] = alphabeta.AlphaBeta{Depth: *player1Option}
	default:
		slog.Error("could not parse -p1 argument", "arg", *player1)
		os.Exit(1)
	}

	switch strings.ToLower(*player2) {
	case "human":
		agents[1] = Human{}
	case "mcts":
		agents[1] = mcts.Mcts{Duration: *player2Option}
	case "minmax":
		agents[1] = minmax.Minmax{Depth: *player2Option}
	case "ab":
		agents[1] = alphabeta.AlphaBeta{Depth: *player2Option}
	default:
		slog.Error("could not parse -p2 argument", "arg", *player2)
		os.Exit(1)
	}

	return agents
}

type Human struct{}

func (h Human) GetMove(p chess.Position) chess.Move {
	fmt.Println("Enter Move (format - s1s2):")
	legalMoves := chess.GenerateLegalMoves(&p)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		move, err := chess.ParseUCIMove(input)
		if err != nil || !slices.Contains(legalMoves, move) {
			fmt.Println("Invalid move")
			continue
		}
		return move
	}
	slog.Error("could not get valid move from human")
	os.Exit(1)
	return chess.Move{}
}

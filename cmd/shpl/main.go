package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/lexer"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/logger"
	"github.com/lorenzobandini/shakespeare-interpreter-go/internal/parser"
)

var debugFlag bool

var rootCmd = &cobra.Command{
	Use:   "shpl",
	Short: "Shakespeare Programming Language interpreter",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debugFlag {
			logger.Init(logger.LevelDebug)
		} else {
			logger.Init(logger.LevelInfo)
		}
	},
}

var tokensCmd = &cobra.Command{
	Use:   "tokens <file>",
	Short: "Tokenize an SPL source file and print the token stream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("cannot read file %q: %w", args[0], err)
		}
		tokens, err := lexer.New(string(src)).ScanTokens()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		out := cmd.OutOrStdout()
		for _, tok := range tokens {
			if _, err := fmt.Fprintln(out, tok.String()); err != nil {
				return err
			}
		}
		return nil
	},
}

var astCmd = &cobra.Command{
	Use:   "ast <file>",
	Short: "Parse an SPL source file and print the AST as JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("cannot read file %q: %w", args[0], err)
		}
		tokens, err := lexer.New(string(src)).ScanTokens()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		prog, err := parser.New(tokens).Parse()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		out, err := json.MarshalIndent(prog, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), string(out)); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug logging")
	rootCmd.AddCommand(tokensCmd, astCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

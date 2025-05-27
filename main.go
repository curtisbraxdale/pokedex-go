package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var commands map[string]cliCommand

func main() {
	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		cleanedInput := cleanInput(input)
		command := cleanedInput[0]
		switch command {
		case "exit":
			commandExit()
		case "help":
			commandHelp()
		default:
			fmt.Printf("Unknown command: %v\n", command)
		}
	}
}

func commandExit() error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	for _, command := range commands {
		fmt.Printf("%s - %s\n", command.name, command.description)
	}
	return nil
}

func cleanInput(text string) []string {
	loweredText := strings.ToLower(text)
	return strings.Fields(loweredText)
}

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

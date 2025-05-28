package main

import (
	"bufio"
	"errors"
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
		"map": {
			name:        "map",
			description: "Displays a list of locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays previous list of locations",
			callback:    commandMapb,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}

	//Initialize the config struct with the first url set
	initialUrl := "https://pokeapi.co/api/v2/location-area/"
	initialPtr := &initialUrl
	config := urlConfig{
		Next:     initialPtr,
		Previous: nil,
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
			commandExit(&config)
		case "help":
			commandHelp(&config)
		case "map":
			commandMap(&config)
		case "mapb":
			commandMapb(&config)
		default:
			fmt.Printf("Unknown command: %v\n", command)
		}
	}
}

func commandExit(config *urlConfig) error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *urlConfig) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	for _, command := range commands {
		fmt.Printf("%s - %s\n", command.name, command.description)
	}
	return nil
}

func commandMap(config *urlConfig) error {
	areas, err := getLocationAreas(config, "forward")
	if err != nil {
		return err
	}
	for i := 0; i < len(areas); i++ {
		fmt.Println(areas[i].Name)
	}
	return nil
}

func commandMapb(config *urlConfig) error {
	if config.Previous != nil {
		areas, err := getLocationAreas(config, "backward")
		if err != nil {
			return err
		}
		for i := 0; i < len(areas); i++ {
			fmt.Println(areas[i].Name)
		}
		return nil
	} else {
		fmt.Println("No previous page available...try again.")
		return errors.New("No previous page available.")
	}
}

func cleanInput(text string) []string {
	loweredText := strings.ToLower(text)
	return strings.Fields(loweredText)
}

type cliCommand struct {
	name        string
	description string
	callback    func(*urlConfig) error
}

type urlConfig struct {
	Previous *string
	Next     *string
}

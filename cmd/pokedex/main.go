package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/curtisbraxdale/pokedex-go/internal/pokecache"
	"github.com/curtisbraxdale/pokedex-go/internal/utils"
)

var dex = utils.Pokedex{Pokemon: make(map[string]utils.Pokemon)}

func main() {
	interval := time.Hour
	cache := pokecache.NewCache(interval)

	//Initialize the config struct with the first url set
	initialUrl := "https://pokeapi.co/api/v2/location-area/"
	initialPtr := &initialUrl
	config := utils.UrlConfig{
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
			commandExit(&config, cache)
		case "help":
			commandHelp(&config, cache)
		case "map":
			commandMap(&config, cache)
		case "mapb":
			commandMapb(&config, cache)
		case "explore":
			commandExplore(cleanedInput[1], &config, cache)
		case "catch":
			commandCatch(cleanedInput[1], &config, cache)
		default:
			fmt.Printf("Unknown command: %v\n", command)
		}
	}
}

func commandExit(config *utils.UrlConfig, cache *pokecache.Cache) error {
	fmt.Print("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *utils.UrlConfig, cache *pokecache.Cache) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	fmt.Printf("help - Displays a help message\n")
	fmt.Printf("map - Displays a list of locations\n")
	fmt.Printf("mapb - Displays previous list of locations\n")
	fmt.Printf("explore - Displays a list of pokemon at a given location\n")
	fmt.Printf("catch - Attempt to catch a given pokemon\n")
	fmt.Printf("exit - Exits the pokedex\n")
	return nil
}

func commandMap(config *utils.UrlConfig, cache *pokecache.Cache) error {
	areas, err := utils.GetLocationAreas(config, "forward", cache)
	if err != nil {
		return err
	}
	for i := 0; i < len(areas); i++ {
		fmt.Println(areas[i].Name)
	}
	return nil
}

func commandMapb(config *utils.UrlConfig, cache *pokecache.Cache) error {
	if config.Previous != nil {
		areas, err := utils.GetLocationAreas(config, "backward", cache)
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

func commandExplore(location string, config *utils.UrlConfig, cache *pokecache.Cache) error {
	pokemonList, err := utils.ExploreArea(location, cache)
	if err != nil {
		return err
	}
	for i := 0; i < len(pokemonList); i++ {
		fmt.Println(pokemonList[i].Pokemon.Name)
	}
	return nil
}

func commandCatch(pokemon string, config *utils.UrlConfig, cache *pokecache.Cache) error {
	pokemonDetails, caught, err := utils.CatchPokemon(pokemon, cache)
	if err != nil {
		fmt.Printf("%s is not a pokemon...try again.\n", pokemon)
		return err
	}
	if caught {
		fmt.Printf("Congratulations! You caught %s!\n", pokemonDetails.Name)
		utils.AddToDex(pokemonDetails, &dex)
	} else {
		fmt.Printf("Oh no! %s escaped!\n", pokemonDetails.Name)
	}
	return nil
}

func cleanInput(text string) []string {
	loweredText := strings.ToLower(text)
	return strings.Fields(loweredText)
}

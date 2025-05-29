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
			commandExit()
		case "help":
			commandHelp()
		case "map":
			commandMap(&config, cache)
		case "mapb":
			commandMapb(&config, cache)
		case "explore":
			commandExplore(cleanedInput[1], cache)
		case "catch":
			commandCatch(cleanedInput[1], cache)
		case "inspect":
			commandInspect(cleanedInput[1])
		case "pokedex":
			commandPokedex()
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
	fmt.Printf("help - Displays a help message\n")
	fmt.Printf("map - Displays a list of locations\n")
	fmt.Printf("mapb - Displays previous list of locations\n")
	fmt.Printf("explore - Displays a list of pokemon at a given location\n")
	fmt.Printf("catch - Attempt to catch a given pokemon\n")
	fmt.Printf("inspect - Shows the pokedex entry of a caught pokemon\n")
	fmt.Printf("pokedex - Displays a list of caught pokemon\n")
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

func commandExplore(location string, cache *pokecache.Cache) error {
	pokemonList, err := utils.ExploreArea(location, cache)
	if err != nil {
		return err
	}
	for i := 0; i < len(pokemonList); i++ {
		fmt.Println(pokemonList[i].Pokemon.Name)
	}
	return nil
}

func commandCatch(pokemon string, cache *pokecache.Cache) error {
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

func commandInspect(pokemon string) error {
	pokemonDetails, err := utils.InspectPokemon(pokemon, &dex)
	if err != nil {
		fmt.Printf("%s is not in your pokedex...try again.\n", pokemon)
		return err
	}
	fmt.Printf("Name: %v\n", pokemonDetails.Name)
	fmt.Printf("Height: %v\n", pokemonDetails.Height)
	fmt.Printf("Weight: %v\n", pokemonDetails.Weight)
	fmt.Printf("Stats:\n")
	fmt.Printf("	-hp: %v\n", utils.GetBaseStat(*pokemonDetails, "hp"))
	fmt.Printf("	-attack: %v\n", utils.GetBaseStat(*pokemonDetails, "attack"))
	fmt.Printf("	-defense: %v\n", utils.GetBaseStat(*pokemonDetails, "defense"))
	fmt.Printf("	-special-attack: %v\n", utils.GetBaseStat(*pokemonDetails, "special-attack"))
	fmt.Printf("	-special-defense: %v\n", utils.GetBaseStat(*pokemonDetails, "special-defense"))
	fmt.Printf("	-speed: %v\n", utils.GetBaseStat(*pokemonDetails, "speed"))
	fmt.Printf("Types:\n")
	fmt.Printf("	-%v\n", utils.GetTypeNames(*pokemonDetails)[0])
	if len(utils.GetTypeNames(*pokemonDetails)) > 1 {
		fmt.Printf("	-%v\n", utils.GetTypeNames(*pokemonDetails)[1])
	}
	return nil
}

func commandPokedex() {
	fmt.Printf("Your Pokedex:\n")
	for _, value := range dex.Pokemon {
		fmt.Printf("	-%v\n", value.Name)
	}
}

func cleanInput(text string) []string {
	loweredText := strings.ToLower(text)
	return strings.Fields(loweredText)
}

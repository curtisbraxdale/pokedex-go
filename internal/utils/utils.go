package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sync"

	"github.com/curtisbraxdale/pokedex-go/internal/pokecache"
)

type UrlConfig struct {
	Previous *string
	Next     *string
}

type Pokedex struct {
	Pokemon map[string]Pokemon
}

// LocationArea represents the structure of a single location area from PokeAPI.
// JSON tags (`json:"name"`) are used to map JSON keys to Go struct fields.
type LocationArea struct {
	ID                   int                   `json:"id"`
	Name                 string                `json:"name"`
	GameIndex            int                   `json:"game_index"`
	EncounterMethodRates []EncounterMethodRate `json:"encounter_method_rates"`
	Location             NamedAPIResource      `json:"location"`
	Names                []Name                `json:"names"`
	PokemonEncounters    []PokemonEncounter    `json:"pokemon_encounters"`
}

type EncounterMethodRate struct {
	EncounterMethod NamedAPIResource `json:"encounter_method"`
	VersionDetails  []VersionDetail  `json:"version_details"`
}

type VersionDetail struct {
	Rate    int              `json:"rate"`
	Version NamedAPIResource `json:"version"`
}

type PokemonEncounter struct {
	Pokemon        NamedAPIResource         `json:"pokemon"`
	VersionDetails []VersionEncounterDetail `json:"version_details"`
}

type VersionEncounterDetail struct {
	EncounterDetails []EncounterDetail `json:"encounter_details"`
	MaxChance        int               `json:"max_chance"`
	Version          NamedAPIResource  `json:"version"`
}

type EncounterDetail struct {
	Chance          int                `json:"chance"`
	ConditionValues []NamedAPIResource `json:"condition_values"`
	MaxLevel        int                `json:"max_level"`
	Method          NamedAPIResource   `json:"method"`
	MinLevel        int                `json:"min_level"`
}

type NamedAPIResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Name represents a localized name for a resource, typically found in 'names' arrays.
type Name struct {
	Name     string           `json:"name"`
	Language NamedAPIResource `json:"language"`
}

type NamedAPIResourceList struct {
	Count    int                `json:"count"`
	Next     *string            `json:"next"`     // Pointer to string as it can be null
	Previous *string            `json:"previous"` // Pointer to string as it can be null
	Results  []NamedAPIResource `json:"results"`
}

// Pokemon represents the main structure of a single Pokemon from PokeAPI.
type Pokemon struct {
	ID             int                `json:"id"`
	Name           string             `json:"name"`
	BaseExperience int                `json:"base_experience"`
	Height         int                `json:"height"`
	IsDefault      bool               `json:"is_default"`
	Order          int                `json:"order"`
	Weight         int                `json:"weight"`
	Abilities      []PokemonAbility   `json:"abilities"`
	Forms          []NamedAPIResource `json:"forms"`
	GameIndices    []VersionGameIndex `json:"game_indices"`
	HeldItems      []PokemonHeldItem  `json:"held_items"`
	Moves          []PokemonMove      `json:"moves"`
	Species        NamedAPIResource   `json:"species"`
	Sprites        PokemonSprites     `json:"sprites"`
	Stats          []PokemonStat      `json:"stats"`
	Types          []PokemonType      `json:"types"`
}

// PokemonAbility represents an ability of a Pokemon.
type PokemonAbility struct {
	IsHidden bool             `json:"is_hidden"`
	Slot     int              `json:"slot"`
	Ability  NamedAPIResource `json:"ability"`
}

// VersionGameIndex represents a game index for a Pokemon.
type VersionGameIndex struct {
	GameIndex int              `json:"game_index"`
	Version   NamedAPIResource `json:"version"`
}

// PokemonHeldItem represents an item held by a Pokemon.
type PokemonHeldItem struct {
	Item           NamedAPIResource         `json:"item"`
	VersionDetails []PokemonHeldItemVersion `json:"version_details"`
}

// PokemonHeldItemVersion represents version details for a held item.
type PokemonHeldItemVersion struct {
	Version NamedAPIResource `json:"version"`
	Rarity  int              `json:"rarity"`
}

// PokemonMove represents a move a Pokemon can learn.
type PokemonMove struct {
	Move                NamedAPIResource     `json:"move"`
	VersionGroupDetails []PokemonMoveVersion `json:"version_group_details"`
}

// PokemonMoveVersion represents version details for a Pokemon move.
type PokemonMoveVersion struct {
	LevelLearnedAt  int              `json:"level_learned_at"`
	MoveLearnMethod NamedAPIResource `json:"move_learn_method"`
	VersionGroup    NamedAPIResource `json:"version_group"`
}

// PokemonSprites contains URLs for various sprites of a Pokemon.
type PokemonSprites struct {
	BackDefault      *string `json:"back_default"`
	BackFemale       *string `json:"back_female"`
	BackShiny        *string `json:"back_shiny"`
	BackShinyFemale  *string `json:"back_shiny_female"`
	FrontDefault     *string `json:"front_default"`
	FrontFemale      *string `json:"front_female"`
	FrontShiny       *string `json:"front_shiny"`
	FrontShinyFemale *string `json:"front_shiny_female"`
	// Add other sprite fields if needed, e.g., "other", "versions"
}

// PokemonStat represents a stat of a Pokemon.
type PokemonStat struct {
	BaseStat int              `json:"base_stat"`
	Effort   int              `json:"effort"`
	Stat     NamedAPIResource `json:"stat"`
}

// PokemonType represents a type of a Pokemon.
type PokemonType struct {
	Slot int              `json:"slot"`
	Type NamedAPIResource `json:"type"`
}

func GetLocationAreas(config *UrlConfig, direction string, cache *pokecache.Cache) ([]LocationArea, error) {
	// Define the API endpoint URL for listing location areas (default limit is 20)
	var listAPIURL string
	if direction == "forward" {
		listAPIURL = *config.Next
	} else {
		listAPIURL = *config.Previous
	}

	fmt.Printf("Fetching list of location areas from: %s\n", listAPIURL)

	// --- Step 1: Fetch the list of NamedAPIResources ---
	// If URL is in Cache, skip Fetch
	data, exists := cache.Get(listAPIURL)
	var resourceList NamedAPIResourceList
	if exists {
		err := json.Unmarshal(data, &resourceList)
		if err != nil {
			fmt.Printf("Error unmarshaling list JSON: %v\n", err)
			fmt.Printf("Response body: %s\n", string(data))
			return nil, err
		}
	} else {
		resp, err := http.Get(listAPIURL)
		if err != nil {
			fmt.Printf("Error making HTTP request for list: %v\n", err)
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Received non-OK HTTP status for list: %s\n", resp.Status)
			return nil, errors.New("non-OK HTTP status")
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading list response body: %v\n", err)
			return nil, err
		}

		// Adds to Cache for later use.
		cache.Add(listAPIURL, body)

		err = json.Unmarshal(body, &resourceList)
		if err != nil {
			fmt.Printf("Error unmarshaling list JSON: %v\n", err)
			fmt.Printf("Response body: %s\n", string(body))
			return nil, err
		}
	}

	// --- Update urlConfig with new next and previous URLs ---
	config.Next = resourceList.Next
	config.Previous = resourceList.Previous

	fmt.Printf("Found %d location areas in the list (showing up to 20 by default).\n", len(resourceList.Results))

	// --- Step 2: Fetch details for each location area concurrently ---
	// Use a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	// Use a channel to collect results safely from goroutines
	locationAreaCh := make(chan LocationArea, len(resourceList.Results))
	// Use a channel to collect errors from goroutines
	errorCh := make(chan error, len(resourceList.Results))

	for _, resource := range resourceList.Results {
		wg.Add(1) // Increment the counter for each goroutine
		go func(url string) {
			defer wg.Done() // Decrement the counter when the goroutine finishes

			//fmt.Printf("  Fetching details for: %s\n", url)
			detailResp, detailErr := http.Get(url)
			if detailErr != nil {
				errorCh <- fmt.Errorf("error making HTTP request for %s: %w", url, detailErr)
				return
			}
			defer detailResp.Body.Close()

			if detailResp.StatusCode != http.StatusOK {
				errorCh <- fmt.Errorf("received non-OK HTTP status for %s: %s", url, detailResp.Status)
				return
			}

			detailBody, detailErr := io.ReadAll(detailResp.Body)
			if detailErr != nil {
				errorCh <- fmt.Errorf("error reading detail response body for %s: %w", url, detailErr)
				return
			}

			var locationArea LocationArea
			detailErr = json.Unmarshal(detailBody, &locationArea)
			if detailErr != nil {
				errorCh <- fmt.Errorf("error unmarshaling detail JSON for %s: %w\nBody: %s", url, detailErr, string(detailBody))
				return
			}
			locationAreaCh <- locationArea // Send the successfully unmarshaled struct to the channel
		}(resource.URL) // Pass the URL to the goroutine
	}

	// Close the channels once all goroutines are done
	go func() {
		wg.Wait()
		close(locationAreaCh)
		close(errorCh)
	}()

	// Collect results and errors
	var allLocationAreas []LocationArea
	var encounteredErrors []error

	for {
		select {
		case la, ok := <-locationAreaCh:
			if !ok { // Channel closed and empty
				locationAreaCh = nil // Mark as nil to stop selecting from it
				break
			}
			allLocationAreas = append(allLocationAreas, la)
		case err, ok := <-errorCh:
			if !ok { // Channel closed and empty
				errorCh = nil // Mark as nil to stop selecting from it
				break
			}
			encounteredErrors = append(encounteredErrors, err)
		}

		// Exit the loop if both channels are closed and empty
		if locationAreaCh == nil && errorCh == nil {
			break
		}
	}

	fmt.Println("\n--- Summary of Fetched Location Areas ---")
	fmt.Printf("Successfully fetched details for %d location areas.\n", len(allLocationAreas))
	fmt.Printf("Encountered %d errors during detail fetching.\n\n", len(encounteredErrors))

	if len(encounteredErrors) > 0 {
		fmt.Println("\n--- Details of Errors ---")
		for _, e := range encounteredErrors {
			fmt.Printf("- %v\n\n", e)
		}
	}

	return allLocationAreas, nil
}

func ExploreArea(location string, cache *pokecache.Cache) ([]PokemonEncounter, error) {
	apiURL := "https://pokeapi.co/api/v2/location-area/" + location + "/"
	fmt.Printf("Exploring area: %s\n", location)

	var locationAreaDetails LocationArea
	var err error
	var body []byte

	// --- Step 1: Fetch the specific LocationArea details ---
	// If URL is in Cache, skip Fetch
	data, exists := cache.Get(apiURL)
	if exists {
		body = data
	} else {
		resp, httpErr := http.Get(apiURL)
		if httpErr != nil {
			return nil, fmt.Errorf("error making HTTP request for location area: %w", httpErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-OK HTTP status for location area: %s", resp.Status)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading location area response body: %w", err)
		}

		cache.Add(apiURL, body)
	}

	// Unmarshal the body (either from cache or HTTP response) into the LocationArea struct
	err = json.Unmarshal(body, &locationAreaDetails)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling location area JSON: %w\nResponse body: %s", err, string(body))
	}
	fmt.Printf("Successfully unmarshaled details for location area: %s (ID: %d)\n", locationAreaDetails.Name, locationAreaDetails.ID)

	// Directly return the PokemonEncounters slice from the fetched LocationArea
	fmt.Printf("Found %d pokemon encounters in the area.\n", len(locationAreaDetails.PokemonEncounters))
	return locationAreaDetails.PokemonEncounters, nil
}

func CatchPokemon(pokemonName string, cache *pokecache.Cache) (*Pokemon, bool, error) {

	apiURL := "https://pokeapi.co/api/v2/pokemon/" + pokemonName + "/"

	var pokemon Pokemon
	var err error
	var body []byte

	// --- Step 1: Fetch the Pokemon ---
	// If URL is in Cache, skip Fetch
	data, exists := cache.Get(apiURL)
	if exists {
		body = data
	} else {
		resp, httpErr := http.Get(apiURL)
		if httpErr != nil {
			return nil, false, fmt.Errorf("error making HTTP request for %s: %w", pokemonName, httpErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, false, fmt.Errorf("received non-OK HTTP status for %s: %s", pokemonName, resp.Status)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, false, fmt.Errorf("error reading response body for %s: %w", pokemonName, err)
		}
		cache.Add(apiURL, body)
	}

	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		fmt.Printf("Error unmarshaling list JSON: %v\n", err)
		fmt.Printf("Response body: %s\n", string(body))
		return nil, false, err
	}

	// --- Simple Catch Logic ---

	// A simple catch rate: higher base_experience makes it harder to catch
	// Let's say, catch if random number (0-100) is greater than base_experience / 2
	catchDifficulty := catchChance(pokemon.BaseExperience, 635, 0.02, 5.0)
	roll := rand.Intn(101) // Random number between 0 and 100
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)
	fmt.Printf("Catch difficulty for %s (Base Exp: %d): %d%%\n", pokemon.Name, pokemon.BaseExperience, catchDifficulty)
	fmt.Printf("Your roll: %d\n", roll)

	if roll > (100 - catchDifficulty) {
		return &pokemon, true, nil
	} else {
		return &pokemon, false, nil
	}
}

func catchChance(baseExp int, maxExp int, minChance float64, k float64) int {
	// Convert inputs to float64 for math.Exp
	normalizedExp := float64(baseExp) / float64(maxExp)
	chance := minChance + (1-minChance)*math.Exp(-k*normalizedExp)

	// Convert to integer percentage (0–100)
	return int(math.Round(chance * 100))
}

func AddToDex(pokemon *Pokemon, pokedex *Pokedex) {
	_, exists := pokedex.Pokemon[pokemon.Name]
	if exists {
		return
	}
	pokedex.Pokemon[pokemon.Name] = *pokemon
	return
}

func InspectPokemon(pokemon string, pokedex *Pokedex) (*Pokemon, error) {
	pokemonDetails, caught := pokedex.Pokemon[pokemon]
	if caught {
		return &pokemonDetails, nil
	}
	return nil, errors.New("Pokemon not in pokedex.")
}

func GetBaseStat(pokemon Pokemon, statName string) int {
	for _, s := range pokemon.Stats {
		if s.Stat.Name == statName {
			return s.BaseStat
		}
	}
	return 0
}

func GetTypeNames(pokemon Pokemon) []string {
	var types []string
	for _, t := range pokemon.Types {
		types = append(types, t.Type.Name)
	}
	return types
}

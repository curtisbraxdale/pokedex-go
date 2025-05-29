package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/curtisbraxdale/pokedex-go/internal/pokecache"
)

type UrlConfig struct {
	Previous *string
	Next     *string
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

func ExploreArea(location string) ([]PokemonEncounter, error) {
	apiURL := "https://pokeapi.co/api/v2/location-area/" + location + "/"
	data, exists := cache.Get(apiURL)
	var resourceList NamedAPIResourceList
	if exists {
		err := json.Unmarshal(data, &resourceList)
		if err != nil {
			fmt.Printf("Error unmarshaling list JSON: %v\n", err)
			fmt.Printf("Response body: %s\n", string(data))
			return nil, err
		}
	} else {
}

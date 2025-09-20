package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/spf13/pflag"

	"github.com/ayanrajpoot10/hianime-api/config"
	"github.com/ayanrajpoot10/hianime-api/internal/api"
	"github.com/ayanrajpoot10/hianime-api/internal/scraper"
)

type App struct {
	scraper *scraper.Scraper
	config  *config.Config
}

func outputJSON(cfg *config.Config, data any) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, output, 0644); err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
		if cfg.Verbose {
			fmt.Printf("Output written to %s\n", cfg.OutputFile)
		}
	} else {
		fmt.Println(string(output))
	}
}

func main() {
	app, command, args := setupApp()

	switch command {
	case "serve", "server", "api":
		app.startAPIServer()
	case "home", "homepage":
		app.scrapHomepage()
	case "search":
		if len(args) < 1 {
			fmt.Println("Usage: hianime search <keyword> [page]")
			fmt.Println("Example: hianime search \"one piece\" 1")
			return
		}
		keyword := args[0]
		page := 1
		if len(args) >= 2 {
			if p, err := strconv.Atoi(args[1]); err == nil {
				page = p
			}
		}
		app.searchAnime(keyword, page)
	case "anime", "details":
		if len(args) < 1 {
			fmt.Println("Usage: hianime anime <anime-id>")
			fmt.Println("Example: hianime anime \"one-piece-100\"")
			return
		}
		animeID := args[0]
		app.getAnimeDetails(animeID)
	case "qtip":
		if len(args) < 1 {
			fmt.Println("Usage: hianime qtip <anime-id>")
			fmt.Println("Example: hianime qtip \"one-piece-100\"")
			return
		}
		animeID := args[0]
		app.getAnimeQtipInfo(animeID)
	case "episodes":
		if len(args) < 1 {
			fmt.Println("Usage: hianime episodes <anime-id>")
			fmt.Println("Example: hianime episodes \"one-piece-100\"")
			return
		}
		animeID := args[0]
		app.getEpisodes(animeID)
	case "list":
		if len(args) < 1 {
			fmt.Println("Usage: hianime list <category> [page]")
			fmt.Println("Example: hianime list \"popular\" 1")
			return
		}
		category := args[0]
		page := 1
		if len(args) >= 2 {
			if p, err := strconv.Atoi(args[1]); err == nil {
				page = p
			}
		}
		app.getAnimeList(category, page)
	case "genre":
		if len(args) < 1 {
			fmt.Println("Usage: hianime genre <genre-name> [page]")
			fmt.Println("Example: hianime genre Action 1")
			return
		}
		genre := args[0]
		page := 1
		if len(args) >= 2 {
			if p, err := strconv.Atoi(args[1]); err == nil {
				page = p
			}
		}
		app.getGenreList(genre, page)
	case "azlist", "az-list":
		if len(args) < 1 {
			fmt.Println("Usage: hianime azlist <sort-option> [page]")
			fmt.Println("Sort options: all, other, A-Z")
			fmt.Println("Example: hianime azlist A 1")
			return
		}
		sortOption := args[0]
		page := 1
		if len(args) >= 2 {
			if p, err := strconv.Atoi(args[1]); err == nil {
				page = p
			}
		}
		app.getAZList(sortOption, page)
	case "servers":
		if len(args) < 1 {
			fmt.Println("Usage: hianime servers <episode-id>")
			fmt.Println("Example: hianime servers \"one-piece-100::ep=2142\"")
			return
		}
		episodeID := args[0]
		app.getServers(episodeID)
	case "stream":
		if len(args) < 3 {
			fmt.Println("Usage: hianime stream <episode-id> <server-type> <server-name>")
			fmt.Println("Example: hianime stream \"one-piece-100::ep=2142\" sub HD-1")
			return
		}
		episodeID := args[0]
		serverType := args[1]
		serverName := args[2]
		app.getStreamLinks(episodeID, serverType, serverName)
	case "suggestions", "suggest":
		if len(args) < 1 {
			fmt.Println("Usage: hianime suggestions <keyword>")
			fmt.Println("Example: hianime suggestions \"naruto\"")
			return
		}
		keyword := args[0]
		app.getSuggestions(keyword)
	case "schedule":
		if len(args) < 1 {
			fmt.Println("Usage: hianime schedule <date> [timezone-offset]")
			fmt.Println("Example: hianime schedule \"2024-01-15\" -330")
			return
		}
		date := args[0]
		tzOffset := -330 // Default to IST
		if len(args) >= 2 {
			if tz, err := strconv.Atoi(args[1]); err == nil {
				tzOffset = tz
			}
		}
		app.getEstimatedSchedule(date, tzOffset)
	case "next-episode", "next":
		if len(args) < 1 {
			fmt.Println("Usage: hianime next-episode <anime-id>")
			fmt.Println("Example: hianime next-episode \"one-piece-100\"")
			return
		}
		animeID := args[0]
		app.getNextEpisodeSchedule(animeID)
	case "producer":
		if len(args) < 1 {
			fmt.Println("Usage: hianime producer <producer-name> [page]")
			fmt.Println("Example: hianime producer \"Ufotable\" 1")
			return
		}
		producerName := args[0]
		page := 1
		if len(args) >= 2 {
			if p, err := strconv.Atoi(args[1]); err == nil && p > 0 {
				page = p
			}
		}
		app.getProducerAnimes(producerName, page)
	case "help", "--help", "-h":
		printUsage()
	case "version", "--version", "-v":
		printVersion()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func setupApp() (*App, string, []string) {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	cfg := config.New()

	pflag.StringVar(&cfg.OutputFile, "output", cfg.OutputFile, "Output file path")
	pflag.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "Enable verbose logging")
	pflag.StringVar(&cfg.Port, "port", cfg.Port, "Port to run the server on")
	pflag.StringVar(&cfg.Host, "host", cfg.Host, "Host to bind the server to")

	pflag.CommandLine.Parse(os.Args[2:])

	args := pflag.Args()

	s := scraper.New(cfg)

	app := &App{
		scraper: s,
		config:  cfg,
	}

	return app, command, args
}

func (a *App) startAPIServer() {
	handler := api.NewHandler(a.scraper)
	router := api.NewRouter(handler, a.config)

	if err := router.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (a *App) scrapHomepage() {
	if a.config.Verbose {
		fmt.Println("Scraping homepage...")
	}

	data, err := a.scraper.Homepage()
	if err != nil {
		log.Fatalf("Failed to scrape homepage: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) searchAnime(keyword string, page int) {
	if a.config.Verbose {
		fmt.Printf("Searching for '%s' (page %d)...\n", keyword, page)
	}

	data, err := a.scraper.Search(keyword, page)
	if err != nil {
		log.Fatalf("Failed to search anime: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getAnimeDetails(animeID string) {
	if a.config.Verbose {
		fmt.Printf("Getting details for anime: %s...\n", animeID)
	}

	data, err := a.scraper.AnimeDetails(animeID)
	if err != nil {
		log.Fatalf("Failed to get anime details: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getAnimeQtipInfo(animeID string) {
	if a.config.Verbose {
		fmt.Printf("Getting qtip info for anime: %s...\n", animeID)
	}

	data, err := a.scraper.GetAnimeQtipInfo(animeID)
	if err != nil {
		log.Fatalf("Failed to get anime qtip info: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getEpisodes(animeID string) {
	if a.config.Verbose {
		fmt.Printf("Getting episodes for anime: %s...\n", animeID)
	}

	data, err := a.scraper.Episodes(animeID)
	if err != nil {
		log.Fatalf("Failed to get episodes: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getAnimeList(category string, page int) {
	if a.config.Verbose {
		fmt.Printf("Getting anime list for category '%s' (page %d)...\n", category, page)
	}

	data, err := a.scraper.AnimeList(category, page)
	if err != nil {
		log.Fatalf("Failed to get anime list: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getGenreList(genre string, page int) {
	if a.config.Verbose {
		fmt.Printf("Getting anime list for genre '%s' (page %d)...\n", genre, page)
	}

	data, err := a.scraper.GenreList(genre, page)
	if err != nil {
		log.Fatalf("Failed to get genre list: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getAZList(sortOption string, page int) {
	if a.config.Verbose {
		fmt.Printf("Getting A-Z list for sort option '%s' (page %d)...\n", sortOption, page)
	}

	data, err := a.scraper.GetAZList(sortOption, page)
	if err != nil {
		log.Fatalf("Failed to get A-Z list: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getProducerAnimes(producerName string, page int) {
	if a.config.Verbose {
		fmt.Printf("Getting animes from producer '%s' (page %d)...\n", producerName, page)
	}

	data, err := a.scraper.GetProducerAnimes(producerName, page)
	if err != nil {
		log.Fatalf("Failed to get producer animes: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getServers(episodeID string) {
	if a.config.Verbose {
		fmt.Printf("Getting servers for episode: %s...\n", episodeID)
	}

	data, err := a.scraper.Servers(episodeID)
	if err != nil {
		log.Fatalf("Failed to get servers: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getStreamLinks(episodeID, serverType, serverName string) {
	if a.config.Verbose {
		fmt.Printf("Getting stream links for episode: %s (type: %s, server: %s)...\n", episodeID, serverType, serverName)
	}

	data, err := a.scraper.StreamLinks(episodeID, serverType, serverName)
	if err != nil {
		log.Fatalf("Failed to get stream links: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getSuggestions(keyword string) {
	if a.config.Verbose {
		fmt.Printf("Getting suggestions for '%s'...\n", keyword)
	}

	data, err := a.scraper.Suggestions(keyword)
	if err != nil {
		log.Fatalf("Failed to get suggestions: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getEstimatedSchedule(date string, tzOffset int) {
	if a.config.Verbose {
		fmt.Printf("Getting estimated schedule for date '%s' (timezone offset: %d)...\n", date, tzOffset)
	}

	data, err := a.scraper.GetEstimatedSchedule(date, tzOffset)
	if err != nil {
		log.Fatalf("Failed to get estimated schedule: %v", err)
	}

	outputJSON(a.config, data)
}

func (a *App) getNextEpisodeSchedule(animeID string) {
	if a.config.Verbose {
		fmt.Printf("Getting next episode schedule for anime: %s...\n", animeID)
	}

	data, err := a.scraper.GetNextEpisodeSchedule(animeID)
	if err != nil {
		log.Fatalf("Failed to get next episode schedule: %v", err)
	}

	outputJSON(a.config, data)
}

func printUsage() {
	fmt.Println(`🎌 HiAnime Scraper CLI

USAGE:
    hianime <COMMAND> [OPTIONS]

COMMANDS:
    serve                          Start the API server
    home                           Scrape homepage content
    search <keyword> [page]        Search for anime
    anime <anime-id>               Get anime details
    qtip <anime-id>                Get anime qtip information
    episodes <anime-id>            Get episode list
    list <category> [page]         Get anime list by category
    genre <genre-name> [page]      Get anime list by genre
    azlist <sort-option> [page]    Get anime list sorted alphabetically (A-Z)
    servers <episode-id>           Get available servers for episode
    stream <episode-id> <type> <server>  Get streaming links for episode
    suggestions <keyword>          Get search suggestions
    schedule <date> [timezone]     Get estimated schedule for date (YYYY-MM-DD)
    next-episode <anime-id>        Get next episode schedule for anime
    producer <producer-name> [page] Get anime list from producer/studio
    help                           Show this help message
    version                        Show version information

OPTIONS:
    --output <file>               Output to file
    --verbose                     Enable verbose logging
    --port <port>                 Server port (default: 3030)
    --host <host>                 Server host (default: 0.0.0.0)

EXAMPLES:
    hianime serve
    hianime home --output home.json
    hianime search "death note" 1
    hianime anime "death-note-60"
    hianime qtip "death-note-60"
    hianime schedule "2025-09-15" -330
    hianime list most-popular 1
    hianime azlist A 1`)
}

func printVersion() {
	fmt.Println("HiAnime Scraper v0.0.1")
}

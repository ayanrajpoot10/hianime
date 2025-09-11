package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"hianime/config"
	"hianime/internal/api"
	"hianime/internal/output"
	"hianime/internal/scraper"
)

type App struct {
	scraper *scraper.Scraper
	config  *config.Config
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
			return
		}
		animeID := args[0]
		app.getAnimeDetails(animeID)
	case "episodes":
		if len(args) < 1 {
			fmt.Println("Usage: hianime episodes <anime-id>")
			return
		}
		animeID := args[0]
		app.getEpisodes(animeID)
	case "list":
		if len(args) < 1 {
			fmt.Println("Usage: hianime list <category> [page]")
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
	case "servers":
		if len(args) < 1 {
			fmt.Println("Usage: hianime servers <episode-id>")
			return
		}
		episodeID := args[0]
		app.getServers(episodeID)
	case "stream":
		if len(args) < 3 {
			fmt.Println("Usage: hianime stream <episode-id> <server-type> <server-name>")
			fmt.Println("Example: hianime stream \"one-piece-100::ep=1\" sub HD-1")
			return
		}
		episodeID := args[0]
		serverType := args[1]
		serverName := args[2]
		app.getStreamLinks(episodeID, serverType, serverName)
	case "suggestions", "suggest":
		if len(args) < 1 {
			fmt.Println("Usage: hianime suggestions <keyword>")
			return
		}
		keyword := args[0]
		app.getSuggestions(keyword)
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

	flagSet := flag.NewFlagSet(command, flag.ExitOnError)

	// Define flags
	format := flagSet.String("format", cfg.OutputFormat, "Output format (json, table, csv)")
	output := flagSet.String("output", cfg.OutputFile, "Output file path")
	verbose := flagSet.Bool("verbose", cfg.Verbose, "Enable verbose logging")
	port := flagSet.String("port", cfg.Port, "Port to run the server on")
	host := flagSet.String("host", cfg.Host, "Host to bind the server to")

	// Parse flags starting from the third argument
	flagSet.Parse(os.Args[2:])

	// Apply parsed flags to config
	cfg.OutputFormat = *format
	cfg.OutputFile = *output
	cfg.Verbose = *verbose
	cfg.Port = *port
	cfg.Host = *host

	args := flagSet.Args()

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

	output.OutputData(a.config, data)
}

func (a *App) searchAnime(keyword string, page int) {
	if a.config.Verbose {
		fmt.Printf("Searching for '%s' (page %d)...\n", keyword, page)
	}

	data, err := a.scraper.Search(keyword, page)
	if err != nil {
		log.Fatalf("Failed to search anime: %v", err)
	}

	output.OutputData(a.config, data)
}

func (a *App) getAnimeDetails(animeID string) {
	if a.config.Verbose {
		fmt.Printf("Getting details for anime: %s...\n", animeID)
	}

	data, err := a.scraper.AnimeDetails(animeID)
	if err != nil {
		log.Fatalf("Failed to get anime details: %v", err)
	}

	output.OutputData(a.config, data)
}

func (a *App) getEpisodes(animeID string) {
	if a.config.Verbose {
		fmt.Printf("Getting episodes for anime: %s...\n", animeID)
	}

	data, err := a.scraper.Episodes(animeID)
	if err != nil {
		log.Fatalf("Failed to get episodes: %v", err)
	}

	output.OutputData(a.config, data)
}

func (a *App) getAnimeList(category string, page int) {
	if a.config.Verbose {
		fmt.Printf("Getting anime list for category '%s' (page %d)...\n", category, page)
	}

	data, err := a.scraper.AnimeList(category, page)
	if err != nil {
		log.Fatalf("Failed to get anime list: %v", err)
	}

	output.OutputData(a.config, data)
}

func (a *App) getGenreList(genre string, page int) {
	if a.config.Verbose {
		fmt.Printf("Getting anime list for genre '%s' (page %d)...\n", genre, page)
	}

	data, err := a.scraper.GenreList(genre, page)
	if err != nil {
		log.Fatalf("Failed to get genre list: %v", err)
	}

	output.OutputData(a.config, data)
}

func (a *App) getServers(episodeID string) {
	if a.config.Verbose {
		fmt.Printf("Getting servers for episode: %s...\n", episodeID)
	}

	data, err := a.scraper.Servers(episodeID)
	if err != nil {
		log.Fatalf("Failed to get servers: %v", err)
	}

	output.OutputData(a.config, data)
}

func (a *App) getStreamLinks(episodeID, serverType, serverName string) {
	if a.config.Verbose {
		fmt.Printf("Getting stream links for episode: %s (type: %s, server: %s)...\n", episodeID, serverType, serverName)
	}

	data, err := a.scraper.StreamLinks(episodeID, serverType, serverName)
	if err != nil {
		log.Fatalf("Failed to get stream links: %v", err)
	}

	output.OutputData(a.config, data)
}

func (a *App) getSuggestions(keyword string) {
	if a.config.Verbose {
		fmt.Printf("Getting suggestions for '%s'...\n", keyword)
	}

	data, err := a.scraper.Suggestions(keyword)
	if err != nil {
		log.Fatalf("Failed to get suggestions: %v", err)
	}

	output.OutputData(a.config, data)
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
    episodes <anime-id>            Get episode list
    list <category> [page]         Get anime list by category
    genre <genre-name> [page]      Get anime list by genre
    servers <episode-id>           Get available servers for episode
    stream <episode-id> <type> <server>  Get streaming links for episode
    suggestions <keyword>          Get search suggestions
    help                           Show this help message
    version                        Show version information

CATEGORIES:
    most-popular, top-airing, most-favorite, completed, recently-added,
    recently-updated, top-upcoming, subbed-anime, dubbed-anime, movie,
    tv, ova, ona, special, events

OPTIONS:
    --format <json|table|csv>     Output format (default: json)
    --output <file>               Output to file
    --verbose                     Enable verbose logging
    --port <port>                 Server port (default: 3030)
    --host <host>                 Server host (default: 0.0.0.0)

EXAMPLES:
    hianime serve
    hianime home --format table
    hianime search "death note" 1
    hianime anime "death-note-60"
    hianime list most-popular 1
    hianime genre action 1 --output anime.csv --format csv`)
}

func printVersion() {
	fmt.Println("HiAnime Scraper v0.0.1")
}

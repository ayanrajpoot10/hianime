package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"text/tabwriter"

	"hianime/config"
	"hianime/internal/api"
	"hianime/internal/scraper"
	"hianime/pkg/models"
)

func main() {
	// Parse command line arguments first
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	// Create config and parse flags for the remaining arguments
	cfg := config.DefaultConfig()
	cfg.LoadFromEnv()

	// Parse flags from the arguments after the command
	flagSet := flag.NewFlagSet(command, flag.ExitOnError)

	// Define flags
	format := flagSet.String("format", cfg.OutputFormat, "Output format (json, table, csv)")
	output := flagSet.String("output", cfg.OutputFile, "Output file path")
	verbose := flagSet.Bool("verbose", cfg.Verbose, "Enable verbose logging")
	port := flagSet.String("port", cfg.Port, "Port to run the server on")
	host := flagSet.String("host", cfg.Host, "Host to bind the server to")

	// Parse flags starting from the second argument
	if len(os.Args) > 2 {
		// Find flags (starting with --)
		var cmdArgs []string
		var flagArgs []string

		for i := 2; i < len(os.Args); i++ {
			if os.Args[i][:1] == "-" {
				flagArgs = append(flagArgs, os.Args[i:]...)
				break
			}
			cmdArgs = append(cmdArgs, os.Args[i])
		}

		if len(flagArgs) > 0 {
			flagSet.Parse(flagArgs)
		}

		// Update os.Args to only include command and non-flag arguments
		os.Args = append([]string{os.Args[0], command}, cmdArgs...)
	}

	// Apply parsed flags to config
	cfg.OutputFormat = *format
	cfg.OutputFile = *output
	cfg.Verbose = *verbose
	cfg.Port = *port
	cfg.Host = *host

	switch command {
	case "serve", "server", "api":
		startAPIServer(cfg)
	case "home", "homepage":
		scrapHomepage(cfg)
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Usage: hianime search <keyword> [page]")
			return
		}
		keyword := os.Args[2]
		page := 1
		if len(os.Args) >= 4 {
			if p, err := strconv.Atoi(os.Args[3]); err == nil {
				page = p
			}
		}
		searchAnime(cfg, keyword, page)
	case "anime", "details":
		if len(os.Args) < 3 {
			fmt.Println("Usage: hianime anime <anime-id>")
			return
		}
		animeID := os.Args[2]
		getAnimeDetails(cfg, animeID)
	case "episodes":
		if len(os.Args) < 3 {
			fmt.Println("Usage: hianime episodes <anime-id>")
			return
		}
		animeID := os.Args[2]
		getEpisodes(cfg, animeID)
	case "list":
		if len(os.Args) < 3 {
			fmt.Println("Usage: hianime list <category> [page]")
			return
		}
		category := os.Args[2]
		page := 1
		if len(os.Args) >= 4 {
			if p, err := strconv.Atoi(os.Args[3]); err == nil {
				page = p
			}
		}
		getAnimeList(cfg, category, page)
	case "genre":
		if len(os.Args) < 3 {
			fmt.Println("Usage: hianime genre <genre-name> [page]")
			return
		}
		genre := os.Args[2]
		page := 1
		if len(os.Args) >= 4 {
			if p, err := strconv.Atoi(os.Args[3]); err == nil {
				page = p
			}
		}
		getGenreList(cfg, genre, page)
	case "servers":
		if len(os.Args) < 3 {
			fmt.Println("Usage: hianime servers <episode-id>")
			return
		}
		episodeID := os.Args[2]
		getServers(cfg, episodeID)
	case "stream":
		if len(os.Args) < 5 {
			fmt.Println("Usage: hianime stream <episode-id> <server-type> <server-name>")
			fmt.Println("Example: hianime stream \"one-piece-100::ep=1\" sub HD-1")
			return
		}
		episodeID := os.Args[2]
		serverType := os.Args[3] // sub or dub
		serverName := os.Args[4] // HD-1, HD-2, etc.
		getStreamLinks(cfg, episodeID, serverType, serverName)
	case "suggestions", "suggest":
		if len(os.Args) < 3 {
			fmt.Println("Usage: hianime suggestions <keyword>")
			return
		}
		keyword := os.Args[2]
		getSuggestions(cfg, keyword)
	case "help", "--help", "-h":
		printUsage()
	case "version", "--version", "-v":
		printVersion()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func startAPIServer(cfg *config.Config) {
	s := scraper.New(cfg)
	handler := api.NewHandler(s)
	router := api.NewRouter(handler, cfg)

	if err := router.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func scrapHomepage(cfg *config.Config) {
	s := scraper.New(cfg)

	if cfg.Verbose {
		fmt.Println("🏠 Scraping homepage...")
	}

	data, err := s.Homepage()
	if err != nil {
		log.Fatalf("Failed to scrape homepage: %v", err)
	}

	outputData(cfg, data)
}

func searchAnime(cfg *config.Config, keyword string, page int) {
	s := scraper.New(cfg)

	if cfg.Verbose {
		fmt.Printf("🔍 Searching for '%s' (page %d)...\n", keyword, page)
	}

	data, err := s.Search(keyword, page)
	if err != nil {
		log.Fatalf("Failed to search anime: %v", err)
	}

	outputData(cfg, data)
}

func getAnimeDetails(cfg *config.Config, animeID string) {
	s := scraper.New(cfg)

	if cfg.Verbose {
		fmt.Printf("📖 Getting details for anime: %s...\n", animeID)
	}

	data, err := s.AnimeDetails(animeID)
	if err != nil {
		log.Fatalf("Failed to get anime details: %v", err)
	}

	outputData(cfg, data)
}

func getEpisodes(cfg *config.Config, animeID string) {
	s := scraper.New(cfg)

	if cfg.Verbose {
		fmt.Printf("📺 Getting episodes for anime: %s...\n", animeID)
	}

	data, err := s.Episodes(animeID)
	if err != nil {
		log.Fatalf("Failed to get episodes: %v", err)
	}

	outputData(cfg, data)
}

func getAnimeList(cfg *config.Config, category string, page int) {
	s := scraper.New(cfg)

	if cfg.Verbose {
		fmt.Printf("📋 Getting anime list for category '%s' (page %d)...\n", category, page)
	}

	data, err := s.AnimeList(category, page)
	if err != nil {
		log.Fatalf("Failed to get anime list: %v", err)
	}

	outputData(cfg, data)
}

func getGenreList(cfg *config.Config, genre string, page int) {
	s := scraper.New(cfg)

	if cfg.Verbose {
		fmt.Printf("🎭 Getting anime list for genre '%s' (page %d)...\n", genre, page)
	}

	data, err := s.GenreList(genre, page)
	if err != nil {
		log.Fatalf("Failed to get genre list: %v", err)
	}

	outputData(cfg, data)
}

func getServers(cfg *config.Config, episodeID string) {
	s := scraper.New(cfg)

	if cfg.Verbose {
		fmt.Printf("🖥️  Getting servers for episode: %s...\n", episodeID)
	}

	data, err := s.Servers(episodeID)
	if err != nil {
		log.Fatalf("Failed to get servers: %v", err)
	}

	outputData(cfg, data)
}

func getStreamLinks(cfg *config.Config, episodeID, serverType, serverName string) {
	s := scraper.New(cfg)

	if cfg.Verbose {
		fmt.Printf("🎬 Getting stream links for episode: %s (type: %s, server: %s)...\n", episodeID, serverType, serverName)
	}

	data, err := s.StreamLinks(episodeID, serverType, serverName)
	if err != nil {
		log.Fatalf("Failed to get stream links: %v", err)
	}

	outputData(cfg, data)
}

func getSuggestions(cfg *config.Config, keyword string) {
	s := scraper.New(cfg)

	if cfg.Verbose {
		fmt.Printf("💡 Getting suggestions for '%s'...\n", keyword)
	}

	data, err := s.Suggestions(keyword)
	if err != nil {
		log.Fatalf("Failed to get suggestions: %v", err)
	}

	outputData(cfg, data)
}

func outputData(cfg *config.Config, data interface{}) {
	switch cfg.OutputFormat {
	case "json":
		outputJSON(cfg, data)
	case "table":
		outputTable(cfg, data)
	case "csv":
		outputCSV(cfg, data)
	default:
		outputJSON(cfg, data)
	}
}

func outputJSON(cfg *config.Config, data interface{}) {
	var output []byte
	var err error

	if cfg.Verbose {
		output, err = json.MarshalIndent(data, "", "  ")
	} else {
		output, err = json.Marshal(data)
	}

	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, output, 0644); err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
		if cfg.Verbose {
			fmt.Printf("✅ Output written to %s\n", cfg.OutputFile)
		}
	} else {
		fmt.Println(string(output))
	}
}

func outputTable(cfg *config.Config, data interface{}) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	switch v := data.(type) {
	case *models.HomepageResponse:
		fmt.Fprintln(w, "TYPE\tRANK\tTITLE\tID\tEPISODES")
		fmt.Fprintln(w, "----\t----\t-----\t--\t--------")

		for _, item := range v.Spotlight {
			fmt.Fprintf(w, "Spotlight\t%d\t%s\t%s\t%d\n", item.Rank, item.Title, item.ID, item.Episodes.Eps)
		}
		for _, item := range v.Trending {
			fmt.Fprintf(w, "Trending\t%d\t%s\t%s\t%d\n", item.Rank, item.Title, item.ID, item.Episodes.Eps)
		}

	case *models.SearchResponse:
		fmt.Fprintln(w, "RANK\tTITLE\tID\tTYPE\tEPISODES")
		fmt.Fprintln(w, "----\t-----\t--\t----\t--------")

		for i, item := range v.Results {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\n", i+1, item.Title, item.ID, item.Type, item.Episodes.Eps)
		}

	case *models.ListPageResponse:
		fmt.Fprintln(w, "RANK\tTITLE\tID\tTYPE\tEPISODES")
		fmt.Fprintln(w, "----\t-----\t--\t----\t--------")

		for i, item := range v.Results {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%d\n", i+1, item.Title, item.ID, item.Type, item.Episodes.Eps)
		}

	case *models.EpisodesResponse:
		fmt.Fprintln(w, "EPISODE\tTITLE\tID\tFILLER")
		fmt.Fprintln(w, "-------\t-----\t--\t------")

		for _, ep := range v.Episodes {
			filler := "No"
			if ep.IsFiller {
				filler = "Yes"
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", ep.Episode, ep.Title, ep.ID, filler)
		}

	case *models.ServersResponse:
		fmt.Fprintln(w, "TYPE\tNAME\tID\tINDEX")
		fmt.Fprintln(w, "----\t----\t--\t-----")

		for _, server := range v.Sub {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", server.Type, server.Name, server.ID, server.Index)
		}
		for _, server := range v.Dub {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", server.Type, server.Name, server.ID, server.Index)
		}

	default:
		// Fallback to JSON for complex types
		outputJSON(cfg, data)
		return
	}

	w.Flush()
}

func outputCSV(cfg *config.Config, data interface{}) {
	var records [][]string

	switch v := data.(type) {
	case *models.HomepageResponse:
		records = append(records, []string{"Type", "Rank", "Title", "ID", "Episodes", "Type"})

		for _, item := range v.Spotlight {
			records = append(records, []string{
				"Spotlight",
				strconv.Itoa(item.Rank),
				item.Title,
				item.ID,
				strconv.Itoa(item.Episodes.Eps),
				item.Type,
			})
		}
		for _, item := range v.Trending {
			records = append(records, []string{
				"Trending",
				strconv.Itoa(item.Rank),
				item.Title,
				item.ID,
				strconv.Itoa(item.Episodes.Eps),
				item.Type,
			})
		}

	case *models.SearchResponse:
		records = append(records, []string{"Rank", "Title", "ID", "Type", "Episodes"})

		for i, item := range v.Results {
			records = append(records, []string{
				strconv.Itoa(i + 1),
				item.Title,
				item.ID,
				item.Type,
				strconv.Itoa(item.Episodes.Eps),
			})
		}

	default:
		// Fallback to JSON for complex types
		outputJSON(cfg, data)
		return
	}

	var writer *csv.Writer
	if cfg.OutputFile != "" {
		file, err := os.Create(cfg.OutputFile)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		defer file.Close()
		writer = csv.NewWriter(file)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}

	defer writer.Flush()

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			log.Fatalf("Failed to write CSV record: %v", err)
		}
	}

	if cfg.OutputFile != "" && cfg.Verbose {
		fmt.Printf("✅ CSV output written to %s\n", cfg.OutputFile)
	}
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

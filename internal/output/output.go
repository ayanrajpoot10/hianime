package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"text/tabwriter"

	"hianime/config"
	"hianime/pkg/models"
)

func OutputData(cfg *config.Config, data any) {
	switch cfg.OutputFormat {
	case "json":
		OutputJSON(cfg, data)
	case "table":
		OutputTable(cfg, data)
	case "csv":
		OutputCSV(cfg, data)
	default:
		OutputJSON(cfg, data)
	}
}

func OutputJSON(cfg *config.Config, data any) {
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
			fmt.Printf("Output written to %s\n", cfg.OutputFile)
		}
	} else {
		fmt.Println(string(output))
	}
}

func OutputTable(cfg *config.Config, data any) {
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
		OutputJSON(cfg, data)
		return
	}

	w.Flush()
}

func OutputCSV(cfg *config.Config, data any) {
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
		OutputJSON(cfg, data)
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
		fmt.Printf("CSV output written to %s\n", cfg.OutputFile)
	}
}

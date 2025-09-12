<p align="center">
  <a href="https://github.com/ayanrajpoot10/hianime">
    <img src="https://raw.githubusercontent.com/Ayanrajpoot10/hianime/refs/heads/main/image/logo.png" alt="HiAnime Logo" width="150" height="150" />
  </a>
</p>

<h1 align="center">HiAnime API</h1>

<p align="center">
  A powerful Go-based scraper for hianime.to, providing both a REST API and CLI for easy data access.
</p>

## ✨ Features

- **REST API Server**: Full-featured HTTP API with JSON responses
- **CLI Tool**: Command-line interface for direct scraping
- **Streaming Links**: Fetch streaming links from multiple servers (not yet implemented)
- **Search & Discovery**: Search anime with suggestions and pagination
- **Comprehensive Data**: Anime details, episodes, servers, and genres
- **Multiple Output Formats**: JSON, Table, and CSV output support
- **Comprehensive Coverage**: All major hianime.to endpoints
- **Rate Limiting**: Built-in request throttling
- **Configurable**: Environment variables and command-line flags
- **CORS Support**: Cross-origin resource sharing for web applications

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- Internet connection

### Installation

1. Clone the repository:
```bash
git clone https://github.com/ayanrajpoot10/hianime
cd hianime
```

2. Build the project:
```bash
# Build CLI tool
go build -o hianime ./cmd/hianime
```

## 📖 Usage

### API Server

Start the API server:
```bash
# Using the main binary
./hianime serve

# Or directly with Go
go run main.go serve

# Custom port
./hianime serve --port 3030
```

The API will be available at `http://localhost:3030` with documentation at the root URL.

### CLI Tool

The CLI tool provides direct access to scraping functions:

```bash
# Get homepage content
./hianime home

# Search for anime
./hianime search "death note"

# Get anime details
./hianime anime "death-note-60"

# Get episode list
./hianime episodes "death-note-60"

# Get anime by category
./hianime list most-popular

# Get anime by genre
./hianime genre action

# Get search suggestions
./hianime suggestions "one piece"

# Get available servers for an episode
./hianime servers "one-piece-100::ep=1"

# Get streaming links (not yet implemented)
./hianime stream "one-piece-100::ep=1" sub HD-1

# Output formats
./hianime home --format table
./hianime search "naruto" --format csv --output results.csv
```

## 🔌 API Endpoints

### Base URL
```
http://localhost:3030/api
```

### Available Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/home` | Homepage content (spotlight, trending, etc.) |
| GET | `/search?keyword={query}&page={page}` | Search anime |
| GET | `/suggestion?keyword={query}` | Search suggestions |
| GET | `/anime/{id}` | Anime details |
| GET | `/episodes/{id}` | Episode list |
| GET | `/animes/{category}?page={page}` | Anime by category |
| GET | `/genre/{genre}?page={page}` | Anime by genre |
| GET | `/servers?id={episodeId}` | Available servers |
| GET | `/stream?id={episodeId}&type={sub\|dub}&server={name}` | **Streaming links (not yet implemented)** |
| GET | `/health` | Health check |


### Example API Requests

```bash
# Get homepage
curl "http://localhost:3030/api/home"

# Search anime
curl "http://localhost:3030/api/search?keyword=death+note&page=1"

# Get anime details
curl "http://localhost:3030/api/anime/death-note-60"

# Get episodes
curl "http://localhost:3030/api/episodes/death-note-60"

# Get most popular anime
curl "http://localhost:3030/api/animes/most-popular?page=1"
```

## ⚙️ Configuration Options

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3030` | Server port |
| `HOST` | `0.0.0.0` | Server host |
| `BASE_URL` | `https://hianime.to` | Target website URL |
| `TIMEOUT` | `30s` | HTTP request timeout |
| `RATE_LIMIT` | `500ms` | Rate limit between requests |
| `OUTPUT_FORMAT` | `json` | Default CLI output format |
| `VERBOSE` | `false` | Enable verbose logging |
| `ENABLE_CORS` | `true` | Enable CORS for API |


## 🔄 API Response Format

All API responses follow this structure:

```json
{
  "success": true,
  "data": {
    // Response data here
  },
  "message": "",
  "error": ""
}
```

### Example Homepage Response

```json
{
  "success": true,
  "data": {
    "spotlight": [
      {
        "id": "one-piece-100",
        "title": "One Piece",
        "alternative_title": "ワンピース",
        "poster": "https://...",
        "rank": 1,
        "type": "TV",
        "episodes": {
          "sub": 1130,
          "dub": 1122,
          "eps": 1130
        }
      }
    ],
    "trending": [...],
    "top10": {
      "today": [...],
      "week": [...],
      "month": [...]
    }
  }
}
```

## ⚠️ Disclaimer

This project is for educational purposes only. It demonstrates web scraping techniques and API development in Go. Please respect the target website's terms of service and implement appropriate rate limiting.

## 📝 License

This project is licensed under the MIT License. See the LICENSE file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 🙏 Acknowledgments

- Built with [goquery](https://github.com/PuerkitoBio/goquery) for HTML parsing
- Inspired by the anime community's need for accessible data

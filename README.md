<p align="center">
  <a href="https://github.com/ayanrajpoot10/hianime">
    <img src="https://raw.githubusercontent.com/Ayanrajpoot10/hianime/refs/heads/main/image/logo.png" alt="HiAnime Logo" width="175" height="175" />
  </a>
</p>

# <p align="center">HiAnime API</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-blue" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
  <img src="https://img.shields.io/github/stars/ayanrajpoot10/hianime?style=social" alt="GitHub Stars">
</p>

<p align="center">
  A powerful Go-based scraper for hianime.to, providing both a REST API and CLI for easy data access.
</p>

## ✨ Features
- **Built with Go** — Fast and efficient web scraper
- **REST API** — JSON endpoints for all major hianime.to features
- **CLI Tool** — Direct scraping from the terminal
- **Streaming Support** — Fetch servers & links (sub/dub)
- **Search & Discovery** — Search, suggestions, categories, genres
- **Rich Data** — Anime details, episodes, schedules, studios
- **Configurable** — Flags & env variables with CORS support
- **Easy Deployment** — One-click deploy to Render or run locally

## 🛠️ Installation

Start using HiAnime API quickly, either locally or in the cloud.

### Install & Run Locally

Install using `go install`:

```bash
go install github.com/ayanrajpoot10/hianime-api/cmd/hianime@latest
```

Check the available commands and options:

```bash
hianime -h
```

---

### Deploy to Render

Deploy your API to Render instantly with a single click:

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/ayanrajpoot10/hianime-api)

**Notes for Render deployment:**

* Render automatically builds your Go project.
* The service will start on default configurations.
* Environment variables can be configured in the Render dashboard as needed.

## 📖 Usage

> **Complete Documentation**: For detailed usage examples, all commands, and extended API reference, see **[USES.md](./USES.md)**

### CLI Tool

The CLI tool provides direct access to scraping functions:

```bash
# Get homepage content
hianime home

# Search for anime (with optional page)
hianime search "death note" 1

# Get anime details
hianime anime "death-note-60"

# Get episode list
hianime episodes "death-note-60"

# Get anime by genre
hianime genre action 1

# Get streaming links (type: sub|dub, server name)
hianime stream "death-note-60::ep=1464" sub HD-2

# Get shedule by date
hianime schedule "2024-01-15" -330
```

### API Server

Start the API server:
```bash
# Using the main binary
hianime serve

# Custom port
hianime serve --port 3030
```

The API will be available at `http://localhost:3030` with documentation at the root URL.

## 🔌 API Endpoints

### Base URL
```
http://localhost:3030/api
```

### Sample testing instance

For quick testing only, you can use the following sample testing instance:

```
https://hianime-api-iy4s.onrender.com
```

Example request against the sample instance:

```bash
curl "https://hianime-api-iy4s.onrender.com/api/home"
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
| GET | `/azlist/{sortOption}?page={page}` | A-Z listing (sort option: A-Z or all) |
| GET | `/servers?id={episodeId}` | Available servers |
| GET | `/stream?id={episodeId}&type={sub\|dub}&server={name}` | **Streaming links** |
| GET | `/schedule?date={YYYY-MM-DD}&tzOffset={offset}` | Estimated schedule for a date |
| GET | `/next-episode/{id}` | Next episode schedule for an anime |
| GET | `/producer/{producer-name}?page={page}` | Anime list by producer/studio |
| GET | `/qtip/{id}` | Short / quick info for an anime |


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

## ⚠️ Disclaimer

This project is for educational purposes only. It demonstrates web scraping techniques and API development in Go. Please respect the target website's terms of service and implement appropriate rate limiting.

## 📝 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 🙏 Acknowledgments

- Built with [goquery](https://github.com/PuerkitoBio/goquery) for HTML parsing
- Ported from the JavaScript project [yahyaMomin/hianime-API](https://github.com/yahyaMomin/hianime-API). This repository ports the original JS project to Go and adds additional features.

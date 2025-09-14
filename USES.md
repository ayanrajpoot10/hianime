# HiAnime Scraper - Complete Usage Guide

## Table of Contents
1. [CLI Commands](#cli-commands)
2. [REST API Endpoints](#rest-api-endpoints)
3. [Response Formats](#response-formats)
4. [Examples](#examples)
5. [Error Handling](#error-handling)
6. [Configuration](#configuration)

---

## CLI Commands

The HiAnime Scraper supports both CLI commands and a REST API server. All CLI commands support output formatting and file output options.

### Global Options

All CLI commands support these global options:

- `--format <json|table|csv>` - Output format (default: json)
- `--output <file>` - Output to file instead of stdout
- `--verbose` - Enable verbose logging
- `--port <port>` - Server port for API mode (default: 3030)
- `--host <host>` - Server host for API mode (default: 0.0.0.0)

### 1. Server Commands

#### Start API Server
```bash
hianime serve [options]
hianime server [options]  # alias
hianime api [options]     # alias
```

**Options:**
- `--port <port>` - Port to run server on (default: 3030)
- `--host <host>` - Host to bind server to (default: 0.0.0.0)

**Examples:**
```bash
# Start server on default port 3030
hianime serve

# Start server on custom port
hianime serve --port 8080

# Start server on specific host and port
hianime serve --host localhost --port 3000
```

### 2. Homepage Commands

#### Get Homepage Content
```bash
hianime home [options]
hianime homepage [options]  # alias
```

**Description:** Scrapes the homepage for trending anime, spotlight content, top rankings, and category lists.

**Examples:**
```bash
# Get homepage in JSON format
hianime home

# Get homepage in table format
hianime home --format table

# Save homepage to file
hianime home --output homepage.json

# Get homepage with verbose logging
hianime home --verbose
```

### 3. Search Commands

#### Search Anime
```bash
hianime search <keyword> [page] [options]
```

**Parameters:**
- `<keyword>` - Search term (required)
- `[page]` - Page number (optional, default: 1)

**Examples:**
```bash
# Search for "death note"
hianime search "death note"

# Search with pagination
hianime search "naruto" 2

# Search and save to CSV
hianime search "one piece" 1 --format csv --output search_results.csv

# Search with table output
hianime search "attack on titan" --format table
```

#### Get Search Suggestions
```bash
hianime suggestions <keyword> [options]
hianime suggest <keyword> [options]  # alias
```

**Parameters:**
- `<keyword>` - Search term for suggestions (required)

**Examples:**
```bash
# Get suggestions for partial search
hianime suggestions "demon"

# Get suggestions in table format
hianime suggestions "dragon" --format table
```

### 4. Anime Details Commands

#### Get Anime Details
```bash
hianime anime <anime-id> [options]
hianime details <anime-id> [options]  # alias
```

**Parameters:**
- `<anime-id>` - Anime ID (e.g., "death-note-60")

**Examples:**
```bash
# Get anime details
hianime anime "death-note-60"

# Get details in table format
hianime anime "one-piece-100" --format table

# Save details to file
hianime anime "naruto-1" --output anime_details.json
```

#### Get Anime Quick Info (Qtip)
```bash
hianime qtip <anime-id> [options]
```

**Parameters:**
- `<anime-id>` - Anime ID (required)

**Description:** Gets quick tooltip information for an anime including basic details and metadata.

**Examples:**
```bash
# Get qtip info
hianime qtip "death-note-60"

# Get qtip info in table format
hianime qtip "one-piece-100" --format table
```

### 5. Episode Commands

#### Get Episode List
```bash
hianime episodes <anime-id> [options]
```

**Parameters:**
- `<anime-id>` - Anime ID (required)

**Examples:**
```bash
# Get episode list
hianime episodes "death-note-60"

# Get episodes in CSV format
hianime episodes "one-piece-100" --format csv --output episodes.csv
```

### 6. Listing Commands

#### Get Anime by Category
```bash
hianime list <category> [page] [options]
```

**Parameters:**
- `<category>` - Category name (required)
- `[page]` - Page number (optional, default: 1)

**Available Categories:**
- `most-popular`
- `top-airing`
- `most-favorite`
- `completed`
- `recently-added`
- `recently-updated`
- `top-upcoming`
- `subbed-anime`
- `dubbed-anime`
- `movie`
- `tv`
- `ova`
- `ona`
- `special`
- `events`

**Examples:**
```bash
# Get most popular anime
hianime list most-popular

# Get top airing anime with pagination
hianime list top-airing 2

# Get movies in table format
hianime list movie --format table

# Save completed anime to CSV
hianime list completed 1 --format csv --output completed_anime.csv
```

#### Get Anime by Genre
```bash
hianime genre <genre-name> [page] [options]
```

**Parameters:**
- `<genre-name>` - Genre name (required)
- `[page]` - Page number (optional, default: 1)

**Examples:**
```bash
# Get action anime
hianime genre action

# Get romance anime with pagination
hianime genre romance 2

# Get comedy anime in table format
hianime genre comedy --format table
```

#### Get A-Z Sorted List
```bash
hianime azlist <sort-option> [page] [options]
hianime az-list <sort-option> [page] [options]  # alias
```

**Parameters:**
- `<sort-option>` - Sort option (required)
- `[page]` - Page number (optional, default: 1)

**Sort Options:**
- `all` - All anime
- `other` - Non-alphabetic titles
- `A` to `Z` - Titles starting with specific letter

**Examples:**
```bash
# Get all anime starting with 'A'
hianime azlist A

# Get anime starting with 'D' on page 2
hianime azlist D 2

# Get all anime in table format
hianime azlist all --format table

# Get other (non-alphabetic) titles
hianime azlist other
```

### 7. Producer/Studio Commands

#### Get Anime by Producer
```bash
hianime producer <producer-name> [page] [options]
```

**Parameters:**
- `<producer-name>` - Producer/studio name (required)
- `[page]` - Page number (optional, default: 1)

**Examples:**
```bash
# Get anime from Studio Ghibli
hianime producer "Studio Ghibli"

# Get anime from Madhouse with pagination
hianime producer "Madhouse" 2

# Get producer anime in table format
hianime producer "Toei Animation" --format table
```

### 8. Streaming Commands

#### Get Available Servers
```bash
hianime servers <episode-id> [options]
```

**Parameters:**
- `<episode-id>` - Episode ID (e.g., "one-piece-100::ep=1")

**Examples:**
```bash
# Get servers for episode
hianime servers "death-note-60::ep=1"

# Get servers in table format
hianime servers "one-piece-100::ep=1" --format table
```

#### Get Stream Links
```bash
hianime stream <episode-id> <server-type> <server-name> [options]
```

**Parameters:**
- `<episode-id>` - Episode ID (required)
- `<server-type>` - Server type: `sub` or `dub` (required)
- `<server-name>` - Server name (e.g., "HD-1", "HD-2") (required)

**Examples:**
```bash
# Get stream links for subbed episode
hianime stream "one-piece-100::ep=1" sub HD-1

# Get stream links for dubbed episode
hianime stream "death-note-60::ep=1" dub HD-2

# Get stream links and save to file
hianime stream "naruto-1::ep=1" sub HD-1 --output stream_links.json
```

### 9. Schedule Commands

#### Get Estimated Schedule
```bash
hianime schedule <date> [timezone-offset] [options]
```

**Parameters:**
- `<date>` - Date in YYYY-MM-DD format (required)
- `[timezone-offset]` - Timezone offset in minutes (optional, default: -330 for IST)

**Examples:**
```bash
# Get schedule for specific date
hianime schedule "2024-01-15"

# Get schedule with custom timezone (UTC)
hianime schedule "2024-01-15" 0

# Get schedule with JST timezone
hianime schedule "2024-01-15" -540

# Get schedule in table format
hianime schedule "2024-01-15" --format table
```

#### Get Next Episode Schedule
```bash
hianime next-episode <anime-id> [options]
hianime next <anime-id> [options]  # alias
```

**Parameters:**
- `<anime-id>` - Anime ID (required)

**Examples:**
```bash
# Get next episode schedule
hianime next-episode "one-piece-100"

# Get next episode info in table format
hianime next "attack-on-titan-112" --format table
```

### 10. Help Commands

#### Show Help
```bash
hianime help
hianime --help
hianime -h
```

#### Show Version
```bash
hianime version
hianime --version
hianime -v
```

---

## REST API Endpoints

The REST API provides the same functionality as the CLI commands through HTTP endpoints. All responses follow a consistent JSON structure.

### Base URL
When running the server: `http://localhost:3030` (default)

### API Response Structure
All API endpoints return responses in this format:
```json
{
  "success": true,
  "data": {...},
  "message": "",
  "error": ""
}
```

### 1. Root Endpoints

#### GET `/`
Returns the API documentation page (HTML).

#### GET `/api`
Returns API information and available endpoints.

**Response:**
```json
{
  "success": true,
  "data": {
    "name": "HiAnime Scraper API",
    "description": "A RESTful API for scraping anime content from hianime.to",
    "endpoints": {...},
    "categories": [...]
  }
}
```

#### GET `/health` or `/api/health`
Health check endpoint.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "message": "hianime API is running"
  }
}
```

### 2. Homepage Endpoints

#### GET `/api/home`
Get homepage content including trending anime, spotlight, and top rankings.

**Response:** [HomepageResponse](#homepage-response)

**Example:**
```bash
curl "http://localhost:3030/api/home"
```

### 3. Search Endpoints

#### GET `/api/search`
Search for anime by keyword.

**Query Parameters:**
- `keyword` (required) - Search term
- `page` (optional) - Page number (default: 1)

**Response:** [SearchResponse](#search-response)

**Examples:**
```bash
curl "http://localhost:3030/api/search?keyword=death%20note"
curl "http://localhost:3030/api/search?keyword=naruto&page=2"
```

#### GET `/api/suggestion`
Get search suggestions for a keyword.

**Query Parameters:**
- `keyword` (required) - Search term

**Response:** Array of suggestion strings

**Example:**
```bash
curl "http://localhost:3030/api/suggestion?keyword=demon"
```

### 4. Anime Details Endpoints

#### GET `/api/anime/{id}`
Get detailed information about a specific anime.

**Path Parameters:**
- `id` (required) - Anime ID (e.g., "death-note-60")

**Response:** [AnimeDetailResponse](#anime-detail-response)

**Example:**
```bash
curl "http://localhost:3030/api/anime/death-note-60"
```

#### GET `/api/qtip/{id}`
Get quick tooltip information for an anime.

**Path Parameters:**
- `id` (required) - Anime ID

**Response:** [QtipResponse](#qtip-response)

**Example:**
```bash
curl "http://localhost:3030/api/qtip/death-note-60"
```

### 5. Episode Endpoints

#### GET `/api/episodes/{id}`
Get episode list for a specific anime.

**Path Parameters:**
- `id` (required) - Anime ID

**Response:** [EpisodesResponse](#episodes-response)

**Example:**
```bash
curl "http://localhost:3030/api/episodes/death-note-60"
```

### 6. Listing Endpoints

#### GET `/api/animes/{category}`
Get anime list by category.

**Path Parameters:**
- `category` (required) - Category name

**Query Parameters:**
- `page` (optional) - Page number (default: 1)

**Response:** [ListPageResponse](#list-page-response)

**Examples:**
```bash
curl "http://localhost:3030/api/animes/most-popular"
curl "http://localhost:3030/api/animes/top-airing?page=2"
```

#### GET `/api/genre/{genre}`
Get anime list by genre.

**Path Parameters:**
- `genre` (required) - Genre name

**Query Parameters:**
- `page` (optional) - Page number (default: 1)

**Response:** [ListPageResponse](#list-page-response)

**Examples:**
```bash
curl "http://localhost:3030/api/genre/action"
curl "http://localhost:3030/api/genre/romance?page=2"
```

#### GET `/api/azlist/{sortOption}`
Get alphabetically sorted anime list.

**Path Parameters:**
- `sortOption` (required) - Sort option (all, other, A-Z)

**Query Parameters:**
- `page` (optional) - Page number (default: 1)

**Response:** [AZListResponse](#azlist-response)

**Examples:**
```bash
curl "http://localhost:3030/api/azlist/A"
curl "http://localhost:3030/api/azlist/all?page=2"
```

### 7. Producer Endpoints

#### GET `/api/producer/{producer-name}`
Get anime list from a specific producer/studio.

**Path Parameters:**
- `producer-name` (required) - Producer/studio name

**Query Parameters:**
- `page` (optional) - Page number (default: 1)

**Response:** [ProducerResponse](#producer-response)

**Examples:**
```bash
curl "http://localhost:3030/api/producer/Studio%20Ghibli"
curl "http://localhost:3030/api/producer/Madhouse?page=2"
```

### 8. Streaming Endpoints

#### GET `/api/servers`
Get available streaming servers for an episode.

**Query Parameters:**
- `id` (required) - Episode ID (e.g., "one-piece-100::ep=1")

**Response:** [ServersResponse](#servers-response)

**Example:**
```bash
curl "http://localhost:3030/api/servers?id=death-note-60::ep=1"
```

#### GET `/api/stream`
Get streaming links for an episode.

**Query Parameters:**
- `id` (required) - Episode ID
- `type` (optional) - Server type: sub/dub (default: sub)
- `server` (optional) - Server name (default: HD-1)

**Response:** [StreamResponse](#stream-response)

**Examples:**
```bash
curl "http://localhost:3030/api/stream?id=one-piece-100::ep=1&type=sub&server=HD-1"
curl "http://localhost:3030/api/stream?id=death-note-60::ep=1&type=dub&server=HD-2"
```

### 9. Schedule Endpoints

#### GET `/api/schedule`
Get estimated schedule for a specific date.

**Query Parameters:**
- `date` (required) - Date in YYYY-MM-DD format
- `tzOffset` (optional) - Timezone offset in minutes (default: -330)

**Response:** [EstimatedScheduleResponse](#estimated-schedule-response)

**Examples:**
```bash
curl "http://localhost:3030/api/schedule?date=2024-01-15"
curl "http://localhost:3030/api/schedule?date=2024-01-15&tzOffset=0"
```

#### GET `/api/next-episode/{id}`
Get next episode schedule for an anime.

**Path Parameters:**
- `id` (required) - Anime ID

**Response:** [NextEpisodeScheduleResponse](#next-episode-schedule-response)

**Example:**
```bash
curl "http://localhost:3030/api/next-episode/one-piece-100"
```

---

## Response Formats

### Homepage Response
```json
{
  "success": true,
  "data": {
    "spotlight": [
      {
        "id": "anime-id",
        "title": "Anime Title",
        "poster": "image-url",
        "description": "Anime description",
        "episodes": {
          "sub": 12,
          "dub": 12,
          "eps": 12
        },
        "type": "TV",
        "rating": "PG-13",
        "quality": "HD"
      }
    ],
    "trending": [...],
    "topAiring": [...],
    "mostPopular": [...],
    "top10": {
      "today": [...],
      "week": [...],
      "month": [...]
    }
  }
}
```

### Search Response
```json
{
  "success": true,
  "data": {
    "results": [
      {
        "id": "anime-id",
        "title": "Anime Title",
        "poster": "image-url",
        "type": "TV",
        "episodes": {
          "sub": 12,
          "dub": 12,
          "eps": 12
        }
      }
    ],
    "hasNextPage": true,
    "currentPage": 1,
    "totalPages": 10
  }
}
```

### Anime Detail Response
```json
{
  "success": true,
  "data": {
    "id": "anime-id",
    "title": "Anime Title",
    "jname": "Japanese Title",
    "poster": "image-url",
    "description": "Detailed description",
    "type": "TV",
    "status": "Completed",
    "aired": "2006 to 2007",
    "episodes": {
      "sub": 37,
      "dub": 37,
      "eps": 37
    },
    "duration": "24m",
    "quality": "HD",
    "rating": "9.0",
    "genres": ["Supernatural", "Thriller", "Psychological"],
    "studios": ["Madhouse"],
    "producers": ["VAP", "Shogakukan-Shueisha Productions"],
    "characters": [...],
    "relatedAnimes": [...],
    "recommendedAnimes": [...],
    "otherSeasons": [...]
  }
}
```

### Episodes Response
```json
{
  "success": true,
  "data": {
    "episodes": [
      {
        "id": "episode-id",
        "title": "Episode Title",
        "episode": 1,
        "url": "episode-url",
        "is_filler": false
      }
    ],
    "totalItems": 37
  }
}
```

### Servers Response
```json
{
  "success": true,
  "data": {
    "episode": 1,
    "sub": [
      {
        "id": "server-id",
        "name": "HD-1",
        "type": "sub",
        "index": 0
      }
    ],
    "dub": [
      {
        "id": "server-id",
        "name": "HD-1",
        "type": "dub",
        "index": 0
      }
    ]
  }
}
```

### Stream Response
```json
{
  "success": true,
  "data": {
    "id": "episode-id",
    "type": "sub",
    "link": {
      "file": "stream-url",
      "type": "hls"
    },
    "tracks": [
      {
        "file": "subtitle-url",
        "kind": "captions"
      }
    ],
    "intro": {
      "start": 0,
      "end": 90
    },
    "outro": {
      "start": 1320,
      "end": 1410
    },
    "server": "HD-1"
  }
}
```

---

## Examples

### CLI Examples

#### Basic Usage
```bash
# Get homepage
hianime home

# Search for anime
hianime search "demon slayer"

# Get anime details
hianime anime "kimetsu-no-yaiba-55"

# Get episodes
hianime episodes "kimetsu-no-yaiba-55"
```

#### Output Formatting
```bash
# JSON output (default)
hianime search "naruto" --format json

# Table output
hianime list most-popular --format table

# CSV output
hianime genre action --format csv --output action_anime.csv
```

#### Advanced Usage
```bash
# Get streaming servers and links
hianime servers "one-piece-100::ep=1"
hianime stream "one-piece-100::ep=1" sub HD-1

# Get schedule information
hianime schedule "2024-01-15" -330
hianime next-episode "one-piece-100"

# Producer-specific anime
hianime producer "Studio Ghibli" 1
```

### API Examples

#### Using curl
```bash
# Get homepage
curl "http://localhost:3030/api/home"

# Search with pagination
curl "http://localhost:3030/api/search?keyword=one%20piece&page=1"

# Get anime details
curl "http://localhost:3030/api/anime/one-piece-100"

# Get streaming links
curl "http://localhost:3030/api/stream?id=one-piece-100::ep=1&type=sub&server=HD-1"
```

#### Using JavaScript (Fetch API)
```javascript
// Search for anime
const searchAnime = async (keyword, page = 1) => {
  const response = await fetch(`http://localhost:3030/api/search?keyword=${encodeURIComponent(keyword)}&page=${page}`);
  const data = await response.json();
  return data;
};

// Get anime details
const getAnimeDetails = async (animeId) => {
  const response = await fetch(`http://localhost:3030/api/anime/${animeId}`);
  const data = await response.json();
  return data;
};

// Get streaming links
const getStreamLinks = async (episodeId, type = 'sub', server = 'HD-1') => {
  const response = await fetch(`http://localhost:3030/api/stream?id=${episodeId}&type=${type}&server=${server}`);
  const data = await response.json();
  return data;
};
```

#### Using Python (requests)
```python
import requests

# Search for anime
def search_anime(keyword, page=1):
    url = f"http://localhost:3030/api/search"
    params = {"keyword": keyword, "page": page}
    response = requests.get(url, params=params)
    return response.json()

# Get anime details
def get_anime_details(anime_id):
    url = f"http://localhost:3030/api/anime/{anime_id}"
    response = requests.get(url)
    return response.json()

# Get homepage
def get_homepage():
    url = "http://localhost:3030/api/home"
    response = requests.get(url)
    return response.json()
```

---

## Error Handling

### CLI Error Handling
The CLI application exits with appropriate error codes and messages:

- **Missing Parameters**: Shows usage information
- **Network Errors**: Shows connection error messages
- **Parsing Errors**: Shows detailed error information
- **File Errors**: Shows file operation error messages

### API Error Handling
API endpoints return error responses in the standard format:

```json
{
  "success": false,
  "data": null,
  "error": "Error description",
  "message": ""
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request (missing parameters)
- `404` - Not Found (invalid endpoint)
- `405` - Method Not Allowed
- `500` - Internal Server Error

**Example Error Response:**
```json
{
  "success": false,
  "data": null,
  "error": "anime not found",
  "message": ""
}
```

---

## Configuration

### Environment Variables
The application can be configured using environment variables:

- `PORT` - Server port (default: 3030)
- `HOST` - Server host (default: 0.0.0.0)
- `BASE_URL` - Base URL for scraping (default: https://hianime.to)
- `TIMEOUT` - HTTP request timeout in seconds (default: 30)
- `RATE_LIMIT` - Rate limit between requests in milliseconds (default: 500)
- `VERBOSE` - Enable verbose logging (default: false)
- `ENABLE_CORS` - Enable CORS headers (default: true)

### Command Line Overrides
CLI flags override environment variables and configuration defaults:

```bash
# Override port and enable verbose mode
hianime serve --port 8080 --verbose

# Override output format and file
hianime search "anime" --format csv --output results.csv
```

### Rate Limiting
The scraper includes built-in rate limiting to avoid overwhelming the target website:

- Default: 500ms between requests
- Configurable via `RATE_LIMIT` environment variable
- Helps prevent IP blocking and ensures stable operation

### CORS Configuration
The API server supports CORS for web applications:

- Enabled by default for all origins
- Configurable via `ENABLE_CORS` environment variable
- Allows cross-origin requests from web browsers

---

## Support

For issues, questions, or contributions, please refer to the project repository and documentation.
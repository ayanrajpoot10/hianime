package models

// AnimeItem represents a single anime item with all possible fields
type AnimeItem struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	JName       string      `json:"jname,omitempty"`
	Poster      string      `json:"poster"`
	Rank        int         `json:"rank,omitempty"`
	Type        string      `json:"type,omitempty"`
	Quality     string      `json:"quality,omitempty"`
	Duration    string      `json:"duration,omitempty"`
	Rating      string      `json:"rating,omitempty"`
	Aired       string      `json:"aired,omitempty"`
	Description string      `json:"description,omitempty"`
	Status      string      `json:"status,omitempty"`
	Episodes    *Episodes   `json:"episodes,omitempty"`
	Characters  []Character `json:"characters,omitempty"`
	Genres      []string    `json:"genres,omitempty"`
	URL         string      `json:"url,omitempty"`
}

// Episodes represents episode information
type Episodes struct {
	Sub int `json:"sub"`
	Dub int `json:"dub"`
	Eps int `json:"eps"`
}

// Character represents anime character information
type Character struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Role    string `json:"role,omitempty"`
}

// Season represents a season of an anime
type Season struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Poster string `json:"poster"`
}

// HomepageResponse represents the response structure for homepage data
type HomepageResponse struct {
	Spotlight       []AnimeItem `json:"spotlight"`
	Trending        []AnimeItem `json:"trending"`
	TopAiring       []AnimeItem `json:"topAiring,omitempty"`
	MostPopular     []AnimeItem `json:"mostPopular,omitempty"`
	MostFavorite    []AnimeItem `json:"mostFavorite,omitempty"`
	LatestCompleted []AnimeItem `json:"latestCompleted,omitempty"`
	LatestUpdated   []AnimeItem `json:"latestUpdated,omitempty"`
	RecentlyAdded   []AnimeItem `json:"recentlyAdded,omitempty"`
	TopUpcoming     []AnimeItem `json:"topUpcoming,omitempty"`
	Top10           Top10       `json:"top10,omitempty"`
	Genres          []string    `json:"genres,omitempty"`
}

// Top10 represents the top 10 anime rankings
type Top10 struct {
	Today []AnimeItem `json:"today,omitempty"`
	Week  []AnimeItem `json:"week,omitempty"`
	Month []AnimeItem `json:"month,omitempty"`
}

// AnimeDetailResponse represents detailed anime information
type AnimeDetailResponse struct {
	AnimeItem
	Studios           []string    `json:"studios,omitempty"`
	Producers         []string    `json:"producers,omitempty"`
	Licensors         []string    `json:"licensors,omitempty"`
	Scored            string      `json:"scored,omitempty"`
	Source            string      `json:"source,omitempty"`
	PremiereDate      string      `json:"premiere_date,omitempty"`
	Synonyms          []string    `json:"synonyms,omitempty"`
	RelatedAnimes     []AnimeItem `json:"related_animes,omitempty"`
	RecommendedAnimes []AnimeItem `json:"recommended_animes,omitempty"`
	OtherSeasons      []Season    `json:"other_seasons,omitempty"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Results     []AnimeItem `json:"results"`
	HasNextPage bool        `json:"hasNextPage"`
	CurrentPage int         `json:"currentPage"`
	TotalPages  int         `json:"totalPages,omitempty"`
}

// EpisodeInfo represents episode information
type EpisodeInfo struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	JName    string `json:"jname,omitempty"`
	URL      string `json:"url"`
	Episode  int    `json:"episode"`
	IsFiller bool   `json:"is_filler"`
}

// EpisodesResponse represents episodes list response
type EpisodesResponse struct {
	Episodes   []EpisodeInfo `json:"episodes"`
	TotalItems int           `json:"totalItems"`
}

// Server represents a streaming server
type Server struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Index int    `json:"index"`
}

// ServersResponse represents available servers for an episode
type ServersResponse struct {
	Episode int      `json:"episode"`
	Sub     []Server `json:"sub"`
	Dub     []Server `json:"dub"`
}

// StreamResponse represents streaming links and sources (matches JS API)
type StreamResponse struct {
	ID     string     `json:"id"`
	Type   string     `json:"type"`
	Link   StreamLink `json:"link"`
	Tracks []Track    `json:"tracks,omitempty"`
	Intro  *TimeRange `json:"intro,omitempty"`
	Outro  *TimeRange `json:"outro,omitempty"`
	Server string     `json:"server"`
	Iframe string     `json:"iframe,omitempty"`
}

// StreamLink represents the main streaming link
type StreamLink struct {
	File string `json:"file"`
	Type string `json:"type"`
}

// Track represents subtitle/thumbnail tracks
type Track struct {
	File string `json:"file"`
	Kind string `json:"kind"`
}

// TimeRange represents intro/outro time ranges
type TimeRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// StreamSource represents a video source (legacy)
type StreamSource struct {
	URL     string `json:"url"`
	Quality string `json:"quality"`
	IsM3U8  bool   `json:"isM3U8"`
}

// GenreResponse represents available genres
type GenreResponse struct {
	Genres []string `json:"genres"`
}

// ListPageResponse represents paginated anime list
type ListPageResponse struct {
	Results     []AnimeItem `json:"results"`
	HasNextPage bool        `json:"hasNextPage"`
	CurrentPage int         `json:"currentPage"`
	Category    string      `json:"category,omitempty"`
}

// QtipAnime represents anime information from qtip endpoint
type QtipAnime struct {
	ID          string    `json:"id"`
	Name        string    `json:"name,omitempty"`
	MalScore    string    `json:"malscore,omitempty"`
	Quality     string    `json:"quality,omitempty"`
	Episodes    *Episodes `json:"episodes,omitempty"`
	Type        string    `json:"type,omitempty"`
	Description string    `json:"description,omitempty"`
	JName       string    `json:"jname,omitempty"`
	Synonyms    string    `json:"synonyms,omitempty"`
	Aired       string    `json:"aired,omitempty"`
	Status      string    `json:"status,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
}

// QtipResponse represents the response structure for qtip data
type QtipResponse struct {
	Anime QtipAnime `json:"anime"`
}

// APIResponse represents a generic API response wrapper
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

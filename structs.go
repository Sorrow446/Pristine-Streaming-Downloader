package main

type Transport struct{}

type WriteCounter struct {
	Total      int64
	TotalStr   string
	Downloaded int64
	Percentage int
	StartTime  int64
}

type Config struct {
	Email    string
	Format   int
	OutPath  string
	Password string
	Urls     []string
}

type Args struct {
	Urls    []string `arg:"positional, required"`
	Format  int      `arg:"-f" default:"-1" help:"Download format.\n\t\t\t 1 = 320 Kbps MP3\n\t\t\t 2 = 16/24-bit FLAC."`
	OutPath string   `arg:"-o" help:"Where to download to. Path will be made if it doesn't already exist."`
}

type AlbumMeta struct {
	Success bool `json:"success"`
	Album   struct {
		ID            int         `json:"id"`
		CreatedAt     string      `json:"created_at"`
		UpdatedAt     string      `json:"updated_at"`
		Code          string      `json:"code"`
		Title         string      `json:"title"`
		Artwork       string      `json:"artwork"`
		Summary       string      `json:"summary"`
		PlaysDaily    int         `json:"plays_daily"`
		PlaysWeekly   int         `json:"plays_weekly"`
		PlaysMonthly  int         `json:"plays_monthly"`
		ImportedAt    string      `json:"imported_at"`
		ExclusiveID   interface{} `json:"exclusive_id"`
		Valid         int         `json:"valid"`
		InvalidReason interface{} `json:"invalid_reason"`
		Tracks        []struct {
			ID        int      `json:"id"`
			CreatedAt string   `json:"created_at"`
			UpdatedAt string   `json:"updated_at"`
			Title     string   `json:"title"`
			Duration  int      `json:"duration"`
			Formats   []string `json:"formats"`
			AlbumID   int      `json:"album_id"`
			Position  int      `json:"position"`
			Sources   struct {
				Mp3  string `json:"mp3"`
				Flac string `json:"flac"`
			} `json:"sources"`
			PlaysDaily   int `json:"plays_daily"`
			PlaysWeekly  int `json:"plays_weekly"`
			PlaysMonthly int `json:"plays_monthly"`
			Album        struct {
				ID            int         `json:"id"`
				CreatedAt     string      `json:"created_at"`
				UpdatedAt     string      `json:"updated_at"`
				Code          string      `json:"code"`
				Title         string      `json:"title"`
				Artwork       string      `json:"artwork"`
				Summary       string      `json:"summary"`
				PlaysDaily    int         `json:"plays_daily"`
				PlaysWeekly   int         `json:"plays_weekly"`
				PlaysMonthly  int         `json:"plays_monthly"`
				ImportedAt    string      `json:"imported_at"`
				ExclusiveID   interface{} `json:"exclusive_id"`
				Valid         int         `json:"valid"`
				InvalidReason interface{} `json:"invalid_reason"`
			} `json:"album"`
		} `json:"tracks"`
	} `json:"album"`
}

type StreamMeta struct {
	Success bool `json:"success"`
	Sources struct {
		Mp3  string `json:"mp3"`
		Flac string `json:"flac"`
	} `json:"sources"`
}

type Format struct {
	Specs     string
	Extension string
	URL       string
}

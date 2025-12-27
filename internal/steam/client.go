package steam

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"steam-fast-api/internal/cache"
	"steam-fast-api/internal/models"
	"strings"
	"time"
	"unicode"
)

type apiResponse struct {
	Response struct {
		Apps            []appEntry `json:"apps"`
		HaveMoreResults bool       `json:"have_more_results"`
		LastAppID       uint32     `json:"last_appid"`
	} `json:"response"`
}

type appEntry struct {
	AppID uint32 `json:"appid"`
	Name  string `json:"name"`
}

func StartScheduler(apiKey string) {
	refresh(apiKey)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			refresh(apiKey)
		}
	}()
}

func refresh(apiKey string) {
	log.Printf("Refreshing Steam database...")
	start := time.Now()
	next := cache.New(250000)
	fetch(apiKey, true, next)
	fetch(apiKey, false, next)
	buildSearchIndex(next)
	cache.Current.Store(next)
	log.Printf("Database updated: %d items in %v", len(next.Names), time.Since(start))
}

func fetch(apiKey string, isGame bool, reg *cache.Registry) {
	baseURL := "https://api.steampowered.com/IStoreService/GetAppList/v1/"
	lastID := uint32(0)
	appType := models.TypeDLC
	gf, df := 0, 1
	if isGame {
		appType = models.TypeGame
		gf, df = 1, 0
	}
	client := &http.Client{Timeout: 30 * time.Second}
	for {
		url := fmt.Sprintf("%s?key=%s&max_results=50000&include_games=%d&include_dlc=%d&last_appid=%d",
			baseURL, apiKey, gf, df, lastID)
		resp, err := client.Get(url)
		if err != nil {
			return
		}
		var data apiResponse
		json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()
		for _, app := range data.Response.Apps {
			reg.Set(app.AppID, app.Name, appType)
		}
		if !data.Response.HaveMoreResults || data.Response.LastAppID == 0 {
			break
		}
		lastID = data.Response.LastAppID
	}
}

func buildSearchIndex(reg *cache.Registry) {
	for id, name := range reg.Names {
		if reg.Types[id]&models.TypeGame == 0 {
			continue
		}
		words := tokenize(name)
		for _, word := range words {
			if len(word) < 2 {
				continue
			}
			reg.Search[word] = append(reg.Search[word], id)
		}
	}
	for _, ids := range reg.Search {
		sort.Slice(ids, func(i, j int) bool {
			return len(reg.Names[ids[i]]) < len(reg.Names[ids[j]])
		})
	}
}

func tokenize(s string) []string {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	return strings.FieldsFunc(strings.ToLower(s), f)
}

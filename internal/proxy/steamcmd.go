package proxy

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"steam-fast-api/internal/models"
	"strconv"
	"strings"
	"time"
)

var client = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	},
}

type AppMetadata struct {
	Name     string
	AppType  models.AppType
	DLCList  []uint32
	ParentID uint32
}

func GetMetadata(appid uint32) (*AppMetadata, error) {
	url := fmt.Sprintf("https://api.steamcmd.net/v1/info/%d", appid)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res models.SteamCMDResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	idStr := strconv.FormatUint(uint64(appid), 10)
	data, ok := res.Data[idStr]
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	meta := &AppMetadata{
		Name:    data.Common.Name,
		AppType: models.TypeGame,
	}

	if strings.ToLower(data.Common.Type) == "dlc" {
		meta.AppType = models.TypeDLC
	}

	if data.Extended.ListOfDLC != "" {
		parts := strings.Split(data.Extended.ListOfDLC, ",")
		for _, p := range parts {
			if id, err := strconv.ParseUint(p, 10, 32); err == nil {
				meta.DLCList = append(meta.DLCList, uint32(id))
			}
		}
	}

	if data.Extended.DLCForAppID != "" {
		if id, err := strconv.ParseUint(data.Extended.DLCForAppID, 10, 32); err == nil {
			meta.ParentID = uint32(id)
		}
	}

	return meta, nil
}

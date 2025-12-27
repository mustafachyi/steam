package models

type AppType byte

const (
	TypeGame AppType = 1 << iota
	TypeDLC
)

type SteamCMDResponse struct {
	Data map[string]struct {
		Common struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"common"`
		Extended struct {
			ListOfDLC   string `json:"listofdlc"`
			DLCForAppID string `json:"dlcforappid"`
		} `json:"extended"`
	} `json:"data"`
}

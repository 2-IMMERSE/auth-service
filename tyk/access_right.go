package tyk

type AccessRight struct {
	ID       string   `json:"api_id"`
	Name     string   `json:"api_name"`
	Versions []string `json:"versions"`
}

package client

type LitePlanStartingCredits struct {
	PackageID int    `json:"package_id"`
	Credits   int    `json:"credits"`
	Criteria  string `json:"criteria"`
}

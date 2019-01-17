package tyk

type Key struct {
	Allowance        int                    `json:"allowance"`
	Rate             int                    `json:"rate"`
	Per              int                    `json:"per"`
	Expires          int                    `json:"expires"`
	QuotaMax         int                    `json:"quota_max"`
	QuotaRenews      int                    `json:"quota_renews"`
	QuotaRemaining   int                    `json:"quota_remaining"`
	QuotaRenewalRate int                    `json:"quota_renewal_rate"`
	Organisation     string                 `json:"org_id"`
	IsInactive       bool                   `json:"is_inactive"`
	AccessRights     map[string]AccessRight `json:"access_rights"`
}

func NewKey(org string) Key {
	// the access rights will be hard coded here for a bit until we have a more dynamic way of handling this
	return Key{
		Allowance:        1000,
		Rate:             1000,
		Per:              60,
		Expires:          -1,
		QuotaMax:         -1,
		QuotaRenews:      86400,
		QuotaRemaining:   0,
		QuotaRenewalRate: 60,
		Organisation:     org,
		IsInactive:       false,
	}
}

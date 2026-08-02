package dashboard

import "time"

type Summary struct {
	TotalMembers        int64 `json:"total_members"`
	TotalCities         int64 `json:"total_cities"`
	TotalCollectedCents int64 `json:"total_collected_cents"`
	ExpectedToDateCents int64 `json:"expected_to_date_cents"`
	OutstandingCents    int64 `json:"outstanding_cents"`
}

type MonthlyRevenue struct {
	Month          time.Time `json:"month"`
	CollectedCents int64     `json:"collected_cents"`
}

type CityRevenue struct {
	CityID         int64  `json:"city_id"`
	CityName       string `json:"city_name"`
	CollectedCents int64  `json:"collected_cents"`
}

type FeeTypeRevenue struct {
	FeeTypeID      int64  `json:"fee_type_id"`
	FeeTypeName    string `json:"fee_type_name"`
	CollectedCents int64  `json:"collected_cents"`
}

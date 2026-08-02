package payment

import "time"

type Payment struct {
	ID             int64     `json:"id"`
	MemberID       int64     `json:"member_id"`
	MemberName     string    `json:"member_name"`
	ChargeID       int64     `json:"charge_id"`
	FeeTypeName    string    `json:"fee_type_name"`
	Period         time.Time `json:"period"`
	AmountCents    int64     `json:"amount_cents"`
	Note           string    `json:"note"`
	RecordedBy     int64     `json:"recorded_by"`
	RecordedByName string    `json:"recorded_by_name"`
	CreatedAt      time.Time `json:"created_at"`
}

type PeriodSummary struct {
	ChargeID          int64     `json:"charge_id"`
	Period            time.Time `json:"period"`
	ChargeAmountCents int64     `json:"charge_amount_cents"`
	PaidCents         int64     `json:"paid_cents"`
	RemainingCents    int64     `json:"remaining_cents"`
	Status            string    `json:"status"`
}

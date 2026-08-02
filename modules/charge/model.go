package charge

import "time"

type MonthlyCharge struct {
	ID          int64      `json:"id"`
	MemberID    int64      `json:"member_id"`
	MemberName  string     `json:"member_name"`
	FeeTypeID   int64      `json:"fee_type_id"`
	FeeTypeName string     `json:"fee_type_name"`
	AmountCents int64      `json:"amount_cents"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"-"`
}

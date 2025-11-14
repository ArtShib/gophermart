package models

import (
	"encoding/json"
	"strconv"
	"time"
)

type Order struct {
	Number     int64   `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual"`
	UploadedAt int64   `json:"uploaded_at"`
	UserID     int64
}

func (o Order) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Number     string  `json:"number"`
		Status     string  `json:"status"`
		Accrual    float64 `json:"accrual"`
		UploadedAt string  `json:"uploaded_at"`
	}{
		Number:     strconv.FormatInt(o.Number, 10),
		Status:     o.Status,
		Accrual:    o.Accrual,
		UploadedAt: time.Unix(o.UploadedAt, 0).Format(time.RFC3339),
	})
}

type OrderArray []Order

//type T struct {
//	Order       string    `json:"order"`
//	Sum         int       `json:"sum"`
//	ProcessedAt time.Time `json:"processed_at"`
//}

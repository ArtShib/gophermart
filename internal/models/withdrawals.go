package models

import (
	"encoding/json"
	"strconv"
	"time"
)

type Withdrawals struct {
	OrderNum    int64   `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt int64   `json:"processed_at"`
}

func (w Withdrawals) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OrderNum    string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}{
		OrderNum:    strconv.FormatInt(w.OrderNum, 10),
		Sum:         w.Sum,
		ProcessedAt: time.Unix(w.ProcessedAt, 0).Format(time.RFC3339),
	})
}

type WithdrawalsArray []Withdrawals

package models

import (
	"encoding/json"
	"strconv"
)

type ResAccrualOrder struct {
	OrderNum int64   `json:"order"`
	Status   string  `json:"status"`
	Accrual  float64 `json:"accrual"`
}

func (r *ResAccrualOrder) UnmarshalJSON(data []byte) error {
	var aux struct {
		OrderNum string  `json:"order"`
		Status   string  `json:"status"`
		Accrual  float64 `json:"accrual"`
	}

	var err error

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	r.Status = aux.Status
	r.Accrual = aux.Accrual

	r.OrderNum, err = strconv.ParseInt(aux.OrderNum, 10, 64)
	if err != nil {
		return err
	}

	return nil
}

type ResAccrualOrderArray []*ResAccrualOrder

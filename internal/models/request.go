package models

import (
	"encoding/json"
	"strconv"
)

type RequestUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RequestWithdraw struct {
	Order int64   `json:"order"`
	Sum   float64 `json:"sum"`
}

func (r *RequestWithdraw) UnmarshalJSON(data []byte) error {
	var aux struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}

	var err error

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	r.Sum = aux.Sum

	r.Order, err = strconv.ParseInt(aux.Order, 10, 64)
	if err != nil {
		return err
	}

	return nil
}

package models

import (
	"encoding/json"
	"time"
)

type Wallet struct {
	Id      string  `json:"id"`
	Balance float64 `json:"balance"`
}

type Transaction struct {
	Id     string      `json:"id"`
	Time   *CustomTime `json:"time"`
	From   string      `json:"from"`
	To     string      `json:"to"`
	Amount float64     `json:"amount"`
}

type CustomTime struct {
	Time time.Time
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	formattedTime := ct.Time.Format("2006-01-02 15:04:05")
	return json.Marshal(formattedTime)
}

package dto

type UpdateExchangeRateRequest struct {
	Pair      string `json:"pair"`
	Rate      string `json:"rate"`
	Timestamp string `json:"timestamp"`
}

type UpdateExchangeRateResponse struct {
	Status string `json:"status"`
}

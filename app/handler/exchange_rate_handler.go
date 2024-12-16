package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"sphere-homework/app/dto"
	"sphere-homework/app/middleware"
	"strconv"
	"strings"
	"time"
)

func ExchangeRateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to parse request", http.StatusBadRequest)
		return
	}

	request := dto.UpdateExchangeRateRequest{}

	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Unable to parse request", http.StatusBadRequest)
		return
	}

	repository := middleware.GetRateRepository(r)

	split := strings.Split(request.Pair, "/")
	if len(split) != 2 {
		http.Error(w, "Invalid pair: "+request.Pair, http.StatusBadRequest)
		return
	}

	rate, err := strconv.ParseFloat(request.Rate, 64)
	if err != nil {
		http.Error(w, "Invalid rate: "+request.Rate, http.StatusBadRequest)
		return
	}

	timestamp, err := time.Parse(time.RFC3339Nano, request.Timestamp)
	if err != nil {
		http.Error(w, "Invalid timestamp: "+request.Timestamp, http.StatusBadRequest)
		return
	}

	err = repository.UpsertRate(split[0], split[1], rate, timestamp)
	if err != nil {
		http.Error(w, "Unable to update rate: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := dto.UpdateExchangeRateResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Unable to write response", http.StatusInternalServerError)
	}
}

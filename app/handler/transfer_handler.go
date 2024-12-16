package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"net/http"
	"sphere-homework/app/dto"
	event2 "sphere-homework/app/event"
	"sphere-homework/app/middleware"
)

func TransferHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	request := dto.TransferRequest{}

	// TBD - do quick balance, user, and asset validation

	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Unable to parse request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	rateRepository := middleware.GetRateRepository(r)

	rate, err := rateRepository.GetRate(request.FromAsset, request.ToAsset)
	if err != nil {
		http.Error(w, "Unable to fetch rate: "+err.Error(), http.StatusBadRequest)
		return
	}

	feeRepository := middleware.GetFeeRepository(r)

	fee, err := feeRepository.GetFee(request.ToAsset)
	if err != nil {
		http.Error(w, "Unable to fetch fee: "+err.Error(), http.StatusBadRequest)
		return
	}

	transferId := uuid.New()
	event, err := event2.NewTransferCreated(request, fee, rate, transferId)
	if err != nil {
		http.Error(w, "Unable to create transfer event", http.StatusBadRequest)
		return
	}

	publisher := middleware.GetEventService(r)
	err = publisher.PublishEvent(*event)
	if err != nil {
		http.Error(w, "Unable publish transfer event", http.StatusBadRequest)
		return
	}

	response := dto.TransferResponse{
		TransferId: transferId,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Unable to write response", http.StatusInternalServerError)
	}
}

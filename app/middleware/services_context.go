package middleware

import (
	"sphere-homework/app/repository"
	"sphere-homework/app/services"
)

type ServicesContext struct {
	EventService     *services.EventService
	RateRepository   *repository.RateRepository
	LedgerRepository *repository.LedgerRepository
	FeeRepository    *repository.FeeRepository
}

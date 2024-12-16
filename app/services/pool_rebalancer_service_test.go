package services

import (
	"github.com/stretchr/testify/assert"
	"sphere-homework/app/model"
	"testing"
)

func TestGetAssetDepositTrending(t *testing.T) {
	balances := []model.LedgerBalance{
		{
			Asset:   "ETH",
			Amount:  10000,
			Inflow:  1000,
			Outflow: 100,
		},
		{
			Asset:   "BTC",
			Amount:  6000,
			Inflow:  10,
			Outflow: 2000,
		},
		{
			Asset:   "CELO",
			Amount:  12000,
			Inflow:  13000,
			Outflow: 1000,
		},
	}
	balance := getAssetDepositMostTrending(balances)

	assert.Equal(t, "CELO", balance.Asset)
}

func TestGetAssetDepositNoTrending1(t *testing.T) {
	balances := []model.LedgerBalance{
		{
			Asset:   "ETH",
			Amount:  10000,
			Inflow:  0,
			Outflow: 100,
		},
		{
			Asset:   "BTC",
			Amount:  6000,
			Inflow:  0,
			Outflow: 2000,
		},
		{
			Asset:   "CELO",
			Amount:  12000,
			Inflow:  0,
			Outflow: 1000,
		},
	}
	balance := getAssetDepositMostTrending(balances)

	assert.Nil(t, balance)
}

func TestGetAssetDepositNoTrending2(t *testing.T) {
	balances := []model.LedgerBalance{
		{
			Asset:   "ETH",
			Amount:  10000,
			Inflow:  20,
			Outflow: 100,
		},
		{
			Asset:   "BTC",
			Amount:  6000,
			Inflow:  30,
			Outflow: 2000,
		},
		{
			Asset:   "CELO",
			Amount:  12000,
			Inflow:  10,
			Outflow: 1000,
		},
	}
	balance := getAssetDepositMostTrending(balances)

	assert.Nil(t, balance)
}

func TestGetAssetDepositTrendingOneElement(t *testing.T) {
	balances := []model.LedgerBalance{
		{
			Asset:   "ETH",
			Amount:  10000,
			Inflow:  200,
			Outflow: 10,
		},
	}
	balance := getAssetDepositMostTrending(balances)

	assert.Equal(t, "ETH", balance.Asset)
}

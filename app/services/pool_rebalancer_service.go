package services

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sphere-homework/app/config"
	"sphere-homework/app/event"
	"sphere-homework/app/model"
	"sphere-homework/app/repository"
	"time"
)

type PoolRebalancerService struct {
	logger              *zap.Logger
	ctx                 context.Context
	config              config.Config
	transferRepository  *repository.TransferRepository
	ledgerRepository    *repository.LedgerRepository
	poolBalancerSetting map[string]PoolReBalancerSetting
	eventService        *EventService
	rateRepository      *repository.RateRepository
}

type PoolReBalancerSetting struct {
	ImbalanceThreshold      float64 // if it exceeds the current threshold, it means we are experiencing some demand (either withdrawals for positive value, or deposits for negative value) and may need to trigger a re-balance if balance is also less than minimum balance
	MinimumBalance          float64 // required minimum balance
	TopUpAmount             float64 // the amount to be added to this pool if a re-balancing is needed
	RequiredBalanceForTopUp float64 // the asset needs have this amount of balance before we transfer out balance from this asset
}

func NewPoolRebalancerService(logger *zap.Logger, ctx context.Context, rateRepository *repository.RateRepository, transferRepository *repository.TransferRepository, ledgerRepository *repository.LedgerRepository, eventService *EventService, config config.Config, poolBalancerSetting map[string]PoolReBalancerSetting) *PoolRebalancerService {
	return &PoolRebalancerService{
		logger:              logger,
		ctx:                 ctx,
		config:              config,
		transferRepository:  transferRepository,
		ledgerRepository:    ledgerRepository,
		eventService:        eventService,
		poolBalancerSetting: poolBalancerSetting,
		rateRepository:      rateRepository,
	}
}

func (p *PoolRebalancerService) Init() {
	go func() {
		p.logger.Info("Starting system pool rebalancer service")

		ticker := time.NewTicker(time.Duration(p.config.PoolRebalancerPollFreqnecySec) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-p.ctx.Done():
				p.logger.Info("Shutting down system rebalancer service")
				return
			case <-ticker.C:
				err := p.checkSystemPool()
				if err != nil {
					p.logger.Error("Unable to check balance")
				}

				return
			}
		}

	}()

}

// algorithm:
// 1. Fetch system balances - and compute inflow and outflow for each balance for a given time duration
// 2. Calculate the imbalance ratio and available liquidity for each system asset
// 3. If an asset's imbalance ratio and minimum required balance exceeds the thresholds configured, find an asset that has the greatest negative imbalance ratio  (meaning this asset has more inflows than the rest) and with balance meeting the minimum required balance
// 4. Execute a re-balance by submitting a transfer request from the asset that has the greatest inflow and meets the minimum balance requirement
func (p *PoolRebalancerService) checkSystemPool() error {
	balances, err := p.ledgerRepository.GetBalances(repository.SystemAccount)

	if err != nil {
		return err
	}

	// This picks an asset pool that has the most incoming deposit ratio as the pool where we will use for funding the re-balancing
	// todo - maybe add more criteria here, like imbalance ratio threshold
	assetTrendingDeposit := getAssetDepositMostTrending(balances)

	p.logger.Info("Checking system pool balances", zap.Any("balances", balances))
	for _, balance := range balances {
		imbalanceRatio := balance.GetImbalanceRatio()
		setting, ok := p.poolBalancerSetting[balance.Asset]

		if !ok {
			continue
		}

		// we need a re-balance here because the asset's loosing liquidity fast, and it is below the minimum required balance
		// note: maybe even if we are not loosing liquidity fast but if below minimum balance, we should also trigger a re-balance?
		if imbalanceRatio >= setting.ImbalanceThreshold && balance.Amount < setting.MinimumBalance {
			if assetTrendingDeposit == nil {
				p.logger.Info("Asset pool needs re-balancing but no source asset pool identified")
				continue
			}

			// ignore potential same asset pool transfers
			if assetTrendingDeposit.Asset == balance.Asset {
				continue
			}

			requiredBalance, ok := p.poolBalancerSetting[assetTrendingDeposit.Asset]
			if !ok {
				p.logger.Info("Unable to find required balance - not re-balancing", zap.Any("asset", assetTrendingDeposit))
				continue
			}

			if assetTrendingDeposit.Amount < requiredBalance.RequiredBalanceForTopUp {
				p.logger.Info("Source pool has less than required balance for re-balancing", zap.Any("asset", assetTrendingDeposit))
				continue
			}

			return p.submitRebalancingTransaction(assetTrendingDeposit, balance.Asset)
		} else {
			p.logger.Info("No need to re-balance asset", zap.Any("balance", balance), zap.Float64("imbalance_ratio", imbalanceRatio))
		}
	}

	return nil
}

func (p *PoolRebalancerService) submitRebalancingTransaction(fromAsset *model.LedgerBalance, toAsset string) error {
	p.logger.Info("Submitting transaction for re-balancing", zap.Any("source_asset", fromAsset), zap.Float64("imbalance_ratio", fromAsset.GetImbalanceRatio()))

	// todo: if we already submitted a re-balancing transaction, don't re-submit it anymore
	topUpAmount := p.poolBalancerSetting[fromAsset.Asset].TopUpAmount

	payload := event.TransferCreated{
		Transfer: event.Transfer{
			TransferId: uuid.New(),
			FromAsset:  fromAsset.Asset,
			ToAsset:    toAsset,
			Sender:     repository.SystemAccount,
			Recipient:  repository.SystemAccount,
			Amount:     topUpAmount,
			Fee:        0,
			Rate:       0,
		},
		Status: "transfer_created",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	baseEvent := event.BaseEvent{
		Timestamp: time.Now().UnixMilli(),
		EventType: "transfer_created",
		Sender:    repository.SystemAccount,
		Payload:   payloadBytes,
	}

	err = p.eventService.PublishEvent(baseEvent)
	if err != nil {
		return err
	}

	return nil
}

func getAssetDepositMostTrending(balances []model.LedgerBalance) *model.LedgerBalance {
	var ret *model.LedgerBalance
	for _, balance := range balances {
		if balance.GetImbalanceRatio() == 0 {
			continue
		}

		// this asset is having more deposits
		if balance.GetImbalanceRatio() < 0 {
			if ret == nil {
				ret = &balance
				continue
			} else {
				if balance.GetImbalanceRatio() < ret.GetImbalanceRatio() {
					ret = &balance
				}
			}
		}
	}

	return ret
}

package main

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"net/http"
	"sphere-homework/app/config"
	"sphere-homework/app/handler"
	"sphere-homework/app/middleware"
	"sphere-homework/app/repository"
	"sphere-homework/app/services"
)

func main() {
	ctx := context.Background()

	// setup logger
	logger, _ := zap.NewProduction()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			fmt.Println("failed to sync zap logger")
		}
	}(logger)

	// setup config
	conf := config.NewConfig()

	// setup kafka producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": conf.KafkaBootstrapServers,
	})
	if err != nil {
		logger.Fatal("failed to create kafka producer", zap.Error(err))
	}
	defer producer.Close()

	// setup consumers
	transferServiceConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": conf.KafkaBootstrapServers,
		"group.id":          "sphere-transfer-service-consumer",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		logger.Fatal("failed to create transfer service kafka consumer", zap.Error(err))
	}
	defer transferServiceConsumer.Close()

	transferHistoryServiceConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": conf.KafkaBootstrapServers,
		"group.id":          "sphere-transfer-history-service-consumer",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		logger.Fatal("failed to create transfer history service kafka consumer", zap.Error(err))
	}
	defer transferHistoryServiceConsumer.Close()

	// setup db
	pool, err := pgxpool.New(context.Background(), conf.DbUrl)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// setup repositories
	transferRepository := repository.NewTransferRepository(pool, ctx)
	exchangeRateRepository := repository.NewRateRepository(pool, ctx, logger)
	ledgerRepository := repository.NewLedgerRepository(pool, ctx, logger)
	feeRepository := repository.NewFeeRepository(pool, ctx)
	transferHistoryRepository := repository.NewTransferHistoryRepository(pool, ctx)

	// setup services

	// hard-code settings for now
	poolBalancerConfig := map[string]services.PoolReBalancerSetting{
		// this means if we are experiencing a lot of withdrawals and our usd amount is less than 10000, then we need to re-balance
		"USD": {
			ImbalanceThreshold: 0.7,
			MinimumBalance:     400000,
			TopUpAmount:        15000,
		},

		"EUR": {
			ImbalanceThreshold: 0.2,
			MinimumBalance:     5000,
			TopUpAmount:        10000,
		},

		"JPY": {
			ImbalanceThreshold: 0.3,
			MinimumBalance:     500000,
			TopUpAmount:        700000,
		},

		"GBP": {
			ImbalanceThreshold: 0.1,
			MinimumBalance:     100000,
			TopUpAmount:        120000,
		},

		"AUD": {
			ImbalanceThreshold: 0.2,
			MinimumBalance:     300000,
			TopUpAmount:        320000,
		},
	}

	eventService := services.NewEventService(producer)
	transferService := services.NewTransferService(transferServiceConsumer, logger, &transferRepository, &ledgerRepository, &eventService, ctx, conf)
	transferHistoryService := services.NewTransferHistoryService(ctx, transferHistoryServiceConsumer, logger, &transferHistoryRepository)
	poolRebalancerService := services.NewPoolRebalancerService(logger, ctx, &exchangeRateRepository, &transferRepository, &ledgerRepository, &eventService, conf, poolBalancerConfig)

	err = transferService.Init()
	if err != nil {
		logger.Fatal("failed to initialize transfer service", zap.Error(err))
	}

	err = transferHistoryService.Init()
	if err != nil {
		logger.Fatal("failed to initialize transfer history service", zap.Error(err))
	}

	poolRebalancerService.Init()

	// setup http handlers
	r := mux.NewRouter()
	r.Use(middleware.InjectorMiddleware(logger, &conf, &middleware.ServicesContext{
		EventService:     &eventService,
		RateRepository:   &exchangeRateRepository,
		LedgerRepository: &ledgerRepository,
		FeeRepository:    &feeRepository,
	}))
	r.Use(middleware.LoggerMiddleware())

	r.HandleFunc("/api/v1/transfer", handler.TransferHandler).Methods("POST")
	r.HandleFunc("/api/v1/exchange-rate", handler.ExchangeRateHandler).Methods("POST")

	logger.Info("Starting sphere transaction server", zap.Int("port", conf.Port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), r); err != nil {
		logger.Fatal("failed to start http server", zap.Error(err))
	}
	logger.Info("Exiting sphere transaction server", zap.Int("port", conf.Port))
}

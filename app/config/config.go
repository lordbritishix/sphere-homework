package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                           int
	DbUrl                          string
	KafkaBootstrapServers          string
	RedisUrl                       string
	TransferOutboxPollFrequencySec int
	PoolRebalancerPollFreqnecySec  int
}

func NewConfig() Config {
	i, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}

	dbUrl := os.Getenv("DB_URL")
	kafkaBootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	redisUrl := os.Getenv("REDIS_URL")
	transferOutboxPollFrequencySec, err := strconv.ParseInt(os.Getenv("TRANSFER_OUTBOX_POLL_FREQUENCY_SEC"), 10, 64)
	if err != nil {
		panic(err)
	}

	poolRebalancerPollFreqnecySec, err := strconv.ParseInt(os.Getenv("POOL_REBALANCER_POLL_FREQUENCY_SEC"), 10, 64)

	return Config{
		Port:                           i,
		DbUrl:                          dbUrl,
		KafkaBootstrapServers:          kafkaBootstrapServers,
		RedisUrl:                       redisUrl,
		TransferOutboxPollFrequencySec: int(transferOutboxPollFrequencySec),
		PoolRebalancerPollFreqnecySec:  int(poolRebalancerPollFreqnecySec),
	}
}

# Architecture

![Architecture](./doc/arch.png)

The transfer service is made up of the following modules that run in their own go-routines (which can easily be run into their own services):
1. Api service - exposes http apis that can be used to initiate transfer or record rates
2. Transfer processor service - manages the transfer request handling and fulfillment. It records transfers in an outbox table. A cron monitors the outbox table and performs the fulfillment. After the transaction request is fulfilled, it is recorded in the ledger which contains the active balance of accounts. The ledger changes are also recorded in the ledger history. 
3. Transfer history service - records transfer events to the transfer history table.
4. Pool balancer service - runs a cron that looks at the system account balances, and performs re-balancing is needed based on a simple algorithm:
   * Fetch system balances - and compute inflow and outflow for each balance for a given time duration
   * Calculate the imbalance ratio and available liquidity for each system asset
   * If an asset's imbalance ratio and minimum required balance exceeds the thresholds configured, find an asset that has the greatest negative imbalance ratio  (meaning this asset has more inflows than the rest) and with balance meeting the minimum required balance
   * Execute a re-balance by submitting a transfer request from the asset that has the greatest inflow and meets the minimum balance requirement

# Pre-requisites
1. Go 1.22.0
2. docker
3. docker-compose

# Notes
1. A `docker-compose` is provided which runs dependent services of the transfer service:
   * Kafka
   * Rates sender
   * Postgres
2. To run the db migration:
   * Install golang-migrate `go get -u github.com/golang-migrate/migrate/v4`
   * Run `migrate -path ./migrations -database 'postgres://postgres:postgres@localhost:5433/sphere_homework?sslmode=disable' up`
3. To run the application:
   * `docker-compose up`
   * `cd app`
   * `go run .`
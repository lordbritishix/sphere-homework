package model

type LedgerBalance struct {
	Asset   string
	Amount  float64
	Inflow  float64
	Outflow float64
}

// GetImbalanceRatio is used to quantify how far a pool’s liquidity flow is from being balanced.
// A positive imbalance ratio means that the pool is loosing liquidity (i.e. more withdrawals are happening)
// A negative imbalance ratio mens that the pool is increasing its liquidity (i.e. more deposits are happening)
func (l *LedgerBalance) GetImbalanceRatio() float64 {
	if l.Amount == 0 {
		return 0
	}

	return (l.Outflow - l.Inflow) / l.Amount
}

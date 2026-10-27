package account

import "errors"

var (
	// ErrInsufficientFunds kastes når avsender ikke har tilstrekkelig saldo.
	ErrInsufficientFunds = errors.New("utilstrekkelig dekning på konto")

	// ErrInvalidAmount kastes hvis beløpet er <= 0, NaN eller uendelig.
	ErrInvalidAmount = errors.New("ugyldig overføringsbeløp (må være et positivt, endelig tall)")

	// ErrSelfTransfer kastes hvis avsender og mottaker er samme konto.
	ErrSelfTransfer = errors.New("overføring til samme konto er ikke tillatt")

	// ErrNilAccount kastes hvis en av kontoene er nil.
	ErrNilAccount = errors.New("konto kan ikke være nil")

	// ErrTimeout kastes hvis overføringen overskrider tidsfristen (context deadline).
	ErrTimeout = errors.New("tidsavbrudd under venting på kontolås")
)

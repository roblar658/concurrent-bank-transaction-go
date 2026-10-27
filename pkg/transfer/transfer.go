package transfer

import (
	"sync/atomic"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/account"
)

// Service definerer grensesnittet for sikker overføring mellom to bankkontoer.
type Service interface {
	Transfer(from, to *account.Account, amount float64) error
}

// OrderedService implementerer deadlock-fri overføring ved deterministisk låseordning.
type OrderedService struct {
	successfulTransfers atomic.Uint64
	failedTransfers     atomic.Uint64
}

// NewOrderedService oppretter en ny instans av OrderedService.
func NewOrderedService() *OrderedService {
	return &OrderedService{}
}

// Transfer utfører en atomisk, deadlock-fri pengeoverføring mellom to kontoer.
//
// Løsningsstrategi:
// 1. Sorterer låsene leksikografisk basert på kontoenes unike ID:
//    Kontoen med lavest ID låses ALLTID først, deretter kontoen med høyest ID.
// 2. Dette eliminerer formelt Coffman-betingelsen om 'Circular Wait'.
// 3. Begge låsene holdes gjennom hele operasjonen (både uttak og innskudd),
//    slik at ingen eksterne lesere eller andre tråder ser en halvveis gjennomført tilstand.
func (s *OrderedService) Transfer(from, to *account.Account, amount float64) error {
	if from == nil || to == nil {
		return account.ErrNilAccount
	}
	if from.ID() == to.ID() {
		return account.ErrSelfTransfer
	}
	if err := account.ValidateAmount(amount); err != nil {
		return err
	}

	// 1. Deterministisk låsesortering basert på unik konto-ID
	first, second := from, to
	if from.ID() > to.ID() {
		first, second = to, from
	}

	// 2. Erverv låser i streng rekkefølge
	first.Lock()
	defer first.Unlock()

	second.Lock()
	defer second.Unlock()

	// 3. Valider dekning mens begge kontoene er trygt låst
	if from.BalanceUnsafe() < amount {
		s.failedTransfers.Add(1)
		return account.ErrInsufficientFunds
	}

	// 4. Atomisk overføring (ingen halvferdig tilstand mulig for omverdenen)
	_ = from.WithdrawUnsafe(amount)
	_ = to.DepositUnsafe(amount)

	s.successfulTransfers.Add(1)
	return nil
}

// SuccessfulCount returnerer antall fullførte overføringer.
func (s *OrderedService) SuccessfulCount() uint64 {
	return s.successfulTransfers.Load()
}

// FailedCount returnerer antall feilede overføringer (f.eks. pga utilstrekkelig dekning).
func (s *OrderedService) FailedCount() uint64 {
	return s.failedTransfers.Load()
}

// ResetStats nullstiller statistikken.
func (s *OrderedService) ResetStats() {
	s.successfulTransfers.Store(0)
	s.failedTransfers.Store(0)
}

// UnorderedService er en naiv implementasjon som låser i vilkårlig rekkefølge (from deretter to).
// ADVARSEL: Eksisterer KUN for demonstrasjon og test av vranglås (deadlock).
type UnorderedService struct{}

func NewUnorderedService() *UnorderedService {
	return &UnorderedService{}
}

func (u *UnorderedService) Transfer(from, to *account.Account, amount float64) error {
	if from == nil || to == nil {
		return account.ErrNilAccount
	}
	if from.ID() == to.ID() {
		return account.ErrSelfTransfer
	}
	if err := account.ValidateAmount(amount); err != nil {
		return err
	}

	// NAIV LÅSING: Låser alltid 'from' først, uansett ID
	from.Lock()
	defer from.Unlock()

	to.Lock()
	defer to.Unlock()

	if from.BalanceUnsafe() < amount {
		return account.ErrInsufficientFunds
	}

	_ = from.WithdrawUnsafe(amount)
	_ = to.DepositUnsafe(amount)
	return nil
}

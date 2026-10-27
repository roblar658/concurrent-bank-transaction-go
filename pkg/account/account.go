package account

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

// Account representerer en trådsikker bankkonto.
// Implementerer sync.Locker for koordinert to-fase låsing.
type Account struct {
	id      string
	owner   string
	balance float64
	mu      sync.Mutex
}

// New oppretter en ny bankkonto med validert ID og positiv startsaldo.
func New(id, owner string, initialBalance float64) (*Account, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return nil, fmt.Errorf("konto-ID kan ikke være tom")
	}
	if initialBalance < 0 || math.IsNaN(initialBalance) || math.IsInf(initialBalance, 0) {
		return nil, fmt.Errorf("ugyldig startsaldo: %v", initialBalance)
	}

	return &Account{
		id:      trimmedID,
		owner:   owner,
		balance: initialBalance,
	}, nil
}

// ID returnerer den unike konto-identifikatoren.
func (a *Account) ID() string {
	return a.id
}

// Owner returnerer kontoeiers navn.
func (a *Account) Owner() string {
	return a.owner
}

// Lock erverver kontoens mutex (sync.Locker).
func (a *Account) Lock() {
	a.mu.Lock()
}

// Unlock frigjør kontoens mutex (sync.Locker).
func (a *Account) Unlock() {
	a.mu.Unlock()
}

// GetBalance leser ut nåværende saldo trådsikkert.
// Låser kontoen slik at ingen observatør kan lese halvferdig tilstand under overføring.
func (a *Account) GetBalance() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

// BalanceUnsafe returnerer saldoen direkte. Forutsetter at låsen allerede holdes!
func (a *Account) BalanceUnsafe() float64 {
	return a.balance
}

// DepositUnsafe setter inn penger uten å ta låsen. Forutsetter at låsen allerede holdes!
func (a *Account) DepositUnsafe(amount float64) error {
	if err := ValidateAmount(amount); err != nil {
		return err
	}
	a.balance += amount
	return nil
}

// WithdrawUnsafe trekker penger uten å ta låsen. Forutsetter at låsen allerede holdes!
func (a *Account) WithdrawUnsafe(amount float64) error {
	if err := ValidateAmount(amount); err != nil {
		return err
	}
	if a.balance < amount {
		return ErrInsufficientFunds
	}
	a.balance -= amount
	return nil
}

// Deposit er et frittstående, trådsikkert innskudd for én konto.
func (a *Account) Deposit(amount float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.DepositUnsafe(amount)
}

// Withdraw er et frittstående, trådsikkert uttak for én konto.
func (a *Account) Withdraw(amount float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.WithdrawUnsafe(amount)
}

// ValidateAmount sjekker om et beløp er et gyldig, positivt og endelig tall.
func ValidateAmount(amount float64) error {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return ErrInvalidAmount
	}
	return nil
}

func (a *Account) String() string {
	return fmt.Sprintf("Account[id=%s, owner=%s, balance=%.2f]", a.id, a.owner, a.GetBalance())
}

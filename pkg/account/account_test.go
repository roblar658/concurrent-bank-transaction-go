package account_test

import (
	"math"
	"testing"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/account"
)

func TestAccountCreation(t *testing.T) {
	acc, err := account.New("ACC-001", "Alice", 1000.0)
	if err != nil {
		t.Fatalf("Uventet feil ved opprettelse: %v", err)
	}
	if acc.ID() != "ACC-001" {
		t.Errorf("Forventet ID ACC-001, fikk %s", acc.ID())
	}
	if acc.Owner() != "Alice" {
		t.Errorf("Forventet eier Alice, fikk %s", acc.Owner())
	}
	if acc.GetBalance() != 1000.0 {
		t.Errorf("Forventet saldo 1000.0, fikk %.2f", acc.GetBalance())
	}

	// Ugyldig tom ID
	if _, err := account.New("   ", "Bob", 100.0); err == nil {
		t.Errorf("Forventet feil ved tom ID, fikk nil")
	}

	// Ugyldig negativ saldo
	if _, err := account.New("ACC-002", "Bob", -10.0); err == nil {
		t.Errorf("Forventet feil ved negativ startsaldo, fikk nil")
	}

	// Ugyldig NaN saldo
	if _, err := account.New("ACC-003", "Bob", math.NaN()); err == nil {
		t.Errorf("Forventet feil ved NaN startsaldo, fikk nil")
	}
}

func TestAccountDepositAndWithdraw(t *testing.T) {
	acc, _ := account.New("ACC-001", "Alice", 500.0)

	// Vellykket innskudd
	if err := acc.Deposit(250.0); err != nil {
		t.Fatalf("Innskudd feilet: %v", err)
	}
	if acc.GetBalance() != 750.0 {
		t.Errorf("Forventet saldo 750.0, fikk %.2f", acc.GetBalance())
	}

	// Vellykket uttak
	if err := acc.Withdraw(150.0); err != nil {
		t.Fatalf("Uttak feilet: %v", err)
	}
	if acc.GetBalance() != 600.0 {
		t.Errorf("Forventet saldo 600.0, fikk %.2f", acc.GetBalance())
	}

	// Overtrekk
	if err := acc.Withdraw(1000.0); err != account.ErrInsufficientFunds {
		t.Errorf("Forventet ErrInsufficientFunds, fikk %v", err)
	}
	if acc.GetBalance() != 600.0 {
		t.Errorf("Saldo skal være uendret etter feilet uttak, fikk %.2f", acc.GetBalance())
	}

	// Ugyldige beløp
	invalidAmounts := []float64{0.0, -10.0, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, inv := range invalidAmounts {
		if err := acc.Deposit(inv); err != account.ErrInvalidAmount {
			t.Errorf("Forventet ErrInvalidAmount for deposit(%v), fikk %v", inv, err)
		}
		if err := acc.Withdraw(inv); err != account.ErrInvalidAmount {
			t.Errorf("Forventet ErrInvalidAmount for withdraw(%v), fikk %v", inv, err)
		}
	}
}

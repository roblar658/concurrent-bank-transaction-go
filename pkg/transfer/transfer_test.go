package transfer_test

import (
	"errors"
	"math"
	"testing"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/account"
	"github.com/roblar658/concurrent-bank-transaction-go/pkg/transfer"
)

func TestTransferSuccess(t *testing.T) {
	accA, _ := account.New("ACC-001", "Alice", 1000.0)
	accB, _ := account.New("ACC-002", "Bob", 500.0)

	svc := transfer.NewOrderedService()

	if err := svc.Transfer(accA, accB, 350.0); err != nil {
		t.Fatalf("Uventet feil under overføring: %v", err)
	}

	if accA.GetBalance() != 650.0 {
		t.Errorf("Forventet 650.0 på konto A, fikk %.2f", accA.GetBalance())
	}
	if accB.GetBalance() != 850.0 {
		t.Errorf("Forventet 850.0 på konto B, fikk %.2f", accB.GetBalance())
	}
	if svc.SuccessfulCount() != 1 {
		t.Errorf("Forventet 1 fullført overføring, fikk %d", svc.SuccessfulCount())
	}
}

func TestTransferFullBalance(t *testing.T) {
	accA, _ := account.New("ACC-001", "Alice", 1000.0)
	accB, _ := account.New("ACC-002", "Bob", 500.0)

	svc := transfer.NewOrderedService()

	if err := svc.Transfer(accA, accB, 1000.0); err != nil {
		t.Fatalf("Overføring av hele saldoen feilet: %v", err)
	}

	if accA.GetBalance() != 0.0 {
		t.Errorf("Forventet 0.0 på konto A, fikk %.2f", accA.GetBalance())
	}
	if accB.GetBalance() != 1500.0 {
		t.Errorf("Forventet 1500.0 på konto B, fikk %.2f", accB.GetBalance())
	}
}

func TestTransferInsufficientFunds(t *testing.T) {
	accA, _ := account.New("ACC-001", "Alice", 1000.0)
	accB, _ := account.New("ACC-002", "Bob", 500.0)

	svc := transfer.NewOrderedService()

	err := svc.Transfer(accA, accB, 1000.01)
	if !errors.Is(err, account.ErrInsufficientFunds) {
		t.Fatalf("Forventet ErrInsufficientFunds, fikk %v", err)
	}

	if accA.GetBalance() != 1000.0 {
		t.Errorf("Avsendersaldo må forbli urørt ved feil, fikk %.2f", accA.GetBalance())
	}
	if accB.GetBalance() != 500.0 {
		t.Errorf("Mottakersaldo må forbli urørt ved feil, fikk %.2f", accB.GetBalance())
	}
	if svc.FailedCount() != 1 {
		t.Errorf("Forventet 1 feilet overføring, fikk %d", svc.FailedCount())
	}
}

func TestTransferInvalidAmounts(t *testing.T) {
	accA, _ := account.New("ACC-001", "Alice", 1000.0)
	accB, _ := account.New("ACC-002", "Bob", 500.0)

	svc := transfer.NewOrderedService()

	invalidAmounts := []float64{0.0, -0.01, -500.0, math.NaN(), math.Inf(1), math.Inf(-1)}
	for _, inv := range invalidAmounts {
		err := svc.Transfer(accA, accB, inv)
		if !errors.Is(err, account.ErrInvalidAmount) {
			t.Errorf("Forventet ErrInvalidAmount for beløp %v, fikk %v", inv, err)
		}
	}
}

func TestTransferNilAccounts(t *testing.T) {
	accA, _ := account.New("ACC-001", "Alice", 1000.0)
	svc := transfer.NewOrderedService()

	if err := svc.Transfer(nil, accA, 100.0); !errors.Is(err, account.ErrNilAccount) {
		t.Errorf("Forventet ErrNilAccount når from er nil, fikk %v", err)
	}
	if err := svc.Transfer(accA, nil, 100.0); !errors.Is(err, account.ErrNilAccount) {
		t.Errorf("Forventet ErrNilAccount når to er nil, fikk %v", err)
	}
	if err := svc.Transfer(nil, nil, 100.0); !errors.Is(err, account.ErrNilAccount) {
		t.Errorf("Forventet ErrNilAccount når begge er nil, fikk %v", err)
	}
}

func TestTransferSelfTransfer(t *testing.T) {
	accA, _ := account.New("ACC-001", "Alice", 1000.0)
	svc := transfer.NewOrderedService()

	if err := svc.Transfer(accA, accA, 100.0); !errors.Is(err, account.ErrSelfTransfer) {
		t.Errorf("Forventet ErrSelfTransfer, fikk %v", err)
	}
}

func TestTransferOrderIndependence(t *testing.T) {
	accA, _ := account.New("ACC-001", "Alice", 1000.0)
	accB, _ := account.New("ACC-002", "Bob", 1000.0)
	svc := transfer.NewOrderedService()

	// Lavere ID -> Høyere ID
	if err := svc.Transfer(accA, accB, 200.0); err != nil {
		t.Fatalf("Feilet: %v", err)
	}
	// Høyere ID -> Lavere ID
	if err := svc.Transfer(accB, accA, 500.0); err != nil {
		t.Fatalf("Feilet: %v", err)
	}

	if accA.GetBalance() != 1300.0 {
		t.Errorf("Forventet 1300.0 for A, fikk %.2f", accA.GetBalance())
	}
	if accB.GetBalance() != 700.0 {
		t.Errorf("Forventet 700.0 for B, fikk %.2f", accB.GetBalance())
	}
}

package transfer_test

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/account"
	"github.com/roblar658/concurrent-bank-transaction-go/pkg/transfer"
)

func TestCircularRingTransfers(t *testing.T) {
	accountCount := 10
	initialPerAccount := 10_000.0
	expectedVault := float64(accountCount) * initialPerAccount

	accounts := make([]*account.Account, accountCount)
	for i := 0; i < accountCount; i++ {
		id := fmt.Sprintf("ACC-%03d", i)
		acc, _ := account.New(id, fmt.Sprintf("User-%d", i), initialPerAccount)
		accounts[i] = acc
	}

	svc := transfer.NewOrderedService()

	goroutines := 20
	transfersPerGoroutine := 1000
	totalExpected := goroutines * transfersPerGoroutine

	startSignal := make(chan struct{})
	var wg sync.WaitGroup

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		gID := g

		go func() {
			defer wg.Done()
			<-startSignal

			for step := 0; step < transfersPerGoroutine; step++ {
				fromIdx := (gID + step) % accountCount
				toIdx := (fromIdx + 1) % accountCount

				_ = svc.Transfer(accounts[fromIdx], accounts[toIdx], 5.0)
			}
		}()
	}

	close(startSignal)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Fullført uten vranglås
	case <-time.After(15 * time.Second):
		t.Fatalf("DEADLOCK DETEKTERT: Sirkulære overføringer i ring hengte seg opp!")
	}

	if svc.SuccessfulCount() != uint64(totalExpected) {
		t.Errorf("Forventet %d overføringer, fikk %d", totalExpected, svc.SuccessfulCount())
	}

	finalVault := 0.0
	for _, acc := range accounts {
		finalVault += acc.GetBalance()
	}

	if math.Abs(finalVault-expectedVault) > 0.0001 {
		t.Errorf("Kapitalen i hvelvet er endret! Forventet %.2f, fikk %.4f", expectedVault, finalVault)
	}
}

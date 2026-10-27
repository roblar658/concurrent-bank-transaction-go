package transfer_test

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/account"
	"github.com/roblar658/concurrent-bank-transaction-go/pkg/transfer"
)

func TestFairnessAndStarvationPrevention(t *testing.T) {
	// Felles flaskehals-konto (oppgjørskonto) som alle goroutiner kjemper om samtidig
	centralAccount, _ := account.New("ACC-CENTRAL", "Central Clearing", 10_000_000.0)

	goroutineCount := 8
	clientAccounts := make([]*account.Account, goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		acc, _ := account.New(fmt.Sprintf("ACC-CLIENT-%d", i), fmt.Sprintf("Client-%d", i), 1_000_000.0)
		clientAccounts[i] = acc
	}

	svc := transfer.NewOrderedService()

	stopSignal := make(chan struct{})
	txCounts := make([]atomic.Uint64, goroutineCount)

	var wg sync.WaitGroup
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		gID := i

		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopSignal:
					return
				default:
					// Kjemper om samme sentralkonto
					_ = svc.Transfer(clientAccounts[gID], centralAccount, 1.0)
					txCounts[gID].Add(1)
					_ = svc.Transfer(centralAccount, clientAccounts[gID], 1.0)
				}
			}
		}()
	}

	// La goroutinene kjempe under høy last i 1.5 sekunder
	time.Sleep(1500 * time.Millisecond)
	close(stopSignal)
	wg.Wait()

	var minTx uint64 = math.MaxUint64
	var maxTx uint64 = 0
	var totalTx uint64 = 0

	for i := 0; i < goroutineCount; i++ {
		c := txCounts[i].Load()
		totalTx += c
		if c < minTx {
			minTx = c
		}
		if c > maxTx {
			maxTx = c
		}
		t.Logf("Goroutine %d gjennomførte %d transaksjoner", i, c)
	}

	t.Logf("Rettferdighetsoversikt: Totalt=%d, Min=%d, Max=%d, Min/Max ratio=%.2f",
		totalTx, minTx, maxTx, float64(minTx)/float64(maxTx))

	// Go sync.Mutex har innebygd 'starvation mode' (kicker inn etter 1ms venting).
	// Ingen goroutine skal være sultet ut:
	if minTx == 0 {
		t.Fatalf("STARVATION DETEKTERT: Minst én goroutine fikk 0 transaksjoner!")
	}

	ratio := float64(minTx) / float64(maxTx)
	if ratio < 0.20 {
		t.Errorf("Ubalansert fordeling (ratio %.2f < 0.20)", ratio)
	}
}

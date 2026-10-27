package transfer_test

import (
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/account"
	"github.com/roblar658/concurrent-bank-transaction-go/pkg/transfer"
)

func TestAtomicityAndIsolation(t *testing.T) {
	initialA := 50_000.0
	initialB := 50_000.0
	expectedTotal := initialA + initialB

	accA, _ := account.New("ACC-ISO-1", "Alice", initialA)
	accB, _ := account.New("ACC-ISO-2", "Bob", initialB)

	svc := transfer.NewOrderedService()

	stopSignal := make(chan struct{})
	var inconsistentObservations atomic.Uint64
	var totalReadObservations atomic.Uint64

	var writersWg sync.WaitGroup
	var readersWg sync.WaitGroup

	// 4 Skriver-goroutiner: Overfører kontinuerlig frem og tilbake
	for i := 0; i < 4; i++ {
		writersWg.Add(1)
		forward := (i%2 == 0)

		go func() {
			defer writersWg.Done()
			for {
				select {
				case <-stopSignal:
					return
				default:
					if forward {
						_ = svc.Transfer(accA, accB, 25.0)
					} else {
						_ = svc.Transfer(accB, accA, 25.0)
					}
				}
			}
		}()
	}

	// 4 Leser-goroutiner: Tar konsistente snapshots ved å låse i ordnet rekkefølge
	for i := 0; i < 4; i++ {
		readersWg.Add(1)

		go func() {
			defer readersWg.Done()
			first, second := accA, accB
			if accA.ID() > accB.ID() {
				first, second = accB, accA
			}

			for {
				select {
				case <-stopSignal:
					return
				default:
					first.Lock()
					second.Lock()

					currentSum := first.BalanceUnsafe() + second.BalanceUnsafe()
					totalReadObservations.Add(1)

					if math.Abs(currentSum-expectedTotal) > 0.0001 {
						inconsistentObservations.Add(1)
					}

					second.Unlock()
					first.Unlock()
				}
			}
		}()
	}

	// Kjør i 1.5 sekunder under full belastning
	time.Sleep(1500 * time.Millisecond)
	close(stopSignal)

	writersWg.Wait()
	readersWg.Wait()

	totalReads := totalReadObservations.Load()
	violations := inconsistentObservations.Load()

	t.Logf("Atomisitetstest fullført: %d konsistente observasjoner gjort.", totalReads)

	if totalReads < 10000 {
		t.Errorf("Forventet minst 10 000 observasjoner under testen, fikk %d", totalReads)
	}

	if violations > 0 {
		t.Fatalf("ISOLASJON BRUTT: %d halvferdige mellomtilstander observert av leserne!", violations)
	}
}

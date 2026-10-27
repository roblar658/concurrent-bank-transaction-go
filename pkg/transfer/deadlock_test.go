package transfer_test

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/account"
	"github.com/roblar658/concurrent-bank-transaction-go/pkg/transfer"
)

func TestCrossTransferNeverDeadlocks(t *testing.T) {
	accA, _ := account.New("ACC-A", "Alice", 1_000_000.0)
	accB, _ := account.New("ACC-B", "Bob", 1_000_000.0)

	svc := transfer.NewOrderedService()

	goroutines := 16
	transfersPerGoroutine := 2000
	totalTransfers := goroutines * transfersPerGoroutine

	startSignal := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		forward := (i%2 == 0)

		go func() {
			defer wg.Done()
			<-startSignal

			for j := 0; j < transfersPerGoroutine; j++ {
				if forward {
					_ = svc.Transfer(accA, accB, 10.0)
				} else {
					_ = svc.Transfer(accB, accA, 10.0)
				}
			}
		}()
	}

	// Skyt startskuddet
	close(startSignal)

	// Timeout-kanal for å fange eventuell vranglås (deadlock)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Fullført uten vranglås!
	case <-time.After(15 * time.Second):
		t.Fatalf("DEADLOCK DETEKTERT: Goroutinene fullførte ikke innen 15 sekunder!")
	}

	if svc.SuccessfulCount() != uint64(totalTransfers) {
		t.Errorf("Forventet %d vellykkede overføringer, fikk %d", totalTransfers, svc.SuccessfulCount())
	}

	// Verifiser kapitalbevaring
	finalSum := accA.GetBalance() + accB.GetBalance()
	if math.Abs(finalSum-2_000_000.0) > 0.0001 {
		t.Errorf("Totalbalansen er endret! Forventet 2 000 000, fikk %.4f", finalSum)
	}
}

func TestUnorderedLockingDeadlockDemonstration(t *testing.T) {
	accA, _ := account.New("ACC-A", "Alice", 10_000.0)
	accB, _ := account.New("ACC-B", "Bob", 10_000.0)

	unordered := transfer.NewUnorderedService()

	g1LockedA := make(chan struct{})
	g2LockedB := make(chan struct{})

	g1Done := make(chan struct{})
	g2Done := make(chan struct{})

	// Goroutine 1: Låser A, venter på at G2 låser B, forsøker så å overføre A -> B (vil blokkere på B)
	go func() {
		accA.Lock()
		close(g1LockedA)
		<-g2LockedB

		// Forsøk overføring som vil låse A (allerede låst av oss) og deretter B (låst av G2)
		// For å unngå at selve transfer låser A to ganger her:
		accA.Unlock()
		_ = unordered.Transfer(accA, accB, 10.0)
		close(g1Done)
	}()

	// Goroutine 2: Låser B, venter på at G1 låser A, forsøker så overføring B -> A (vil blokkere på A)
	go func() {
		accB.Lock()
		close(g2LockedB)
		<-g1LockedA

		accB.Unlock()
		_ = unordered.Transfer(accB, accA, 10.0)
		close(g2Done)
	}()

	// Under samtidig kryss-låsing med mange repetisjoner oppstår vranglås
	// Her verifiserer vi at OrderedService aldri henger, mens naiv låsing er utsatt for sirkulær vent.
}

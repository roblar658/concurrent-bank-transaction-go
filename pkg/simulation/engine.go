package simulation

import (
	"errors"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/account"
	"github.com/roblar658/concurrent-bank-transaction-go/pkg/transfer"
)

// Engine orkestrerer parallelle overføringssimuleringer med sanntidsrevisjon i Go.
type Engine struct {
	service transfer.Service
}

func NewEngine(service transfer.Service) *Engine {
	return &Engine{service: service}
}

// RunSimulation kjører transaksjonene fordelt over N goroutiner og logger metrikker.
func (e *Engine) RunSimulation(
	accounts map[string]*account.Account,
	transactions []TransactionRecord,
	workers int,
	enableAudit bool,
) *SimulationMetrics {
	if workers <= 0 {
		workers = 4
	}

	initialVault := calculateTotalVault(accounts)
	metrics := NewMetrics()

	stopAuditor := make(chan struct{})
	var auditorWg sync.WaitGroup

	if enableAudit {
		orderedAccounts := make([]*account.Account, 0, len(accounts))
		for _, acc := range accounts {
			orderedAccounts = append(orderedAccounts, acc)
		}
		sort.Slice(orderedAccounts, func(i, j int) bool {
			return orderedAccounts[i].ID() < orderedAccounts[j].ID()
		})

		auditorWg.Add(1)
		go func() {
			defer auditorWg.Done()
			ticker := time.NewTicker(15 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-stopAuditor:
					return
				case <-ticker.C:
					// Snapshot-lås alle kontoer deterministisk
					auditedSum := snapshotGlobal(orderedAccounts, 0)
					consistent := math.Abs(auditedSum-initialVault) < 0.0001
					metrics.RecordAudit(consistent)
				}
			}
		}()
	}

	// Opprett transaksjonskanal for arbeidsfordeling
	jobs := make(chan TransactionRecord, len(transactions))
	for _, tx := range transactions {
		jobs <- tx
	}
	close(jobs)

	var workersWg sync.WaitGroup
	for w := 0; w < workers; w++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()
			for tx := range jobs {
				from := accounts[tx.FromID]
				to := accounts[tx.ToID]

				if from == nil || to == nil {
					metrics.RecordMissingAccount()
					continue
				}

				start := time.Now()
				err := e.service.Transfer(from, to, tx.Amount)
				elapsed := time.Since(start)

				if err == nil {
					metrics.RecordSuccess(elapsed)
				} else if errors.Is(err, account.ErrInsufficientFunds) {
					metrics.RecordInsufficientFunds(elapsed)
				} else if errors.Is(err, account.ErrSelfTransfer) {
					metrics.RecordSelfTransfer()
				} else if errors.Is(err, account.ErrInvalidAmount) {
					metrics.RecordInvalidAmount()
				}
			}
		}()
	}

	workersWg.Wait()

	if enableAudit {
		close(stopAuditor)
		auditorWg.Wait()
	}

	finalVault := calculateTotalVault(accounts)
	metrics.Complete(initialVault, finalVault)
	return metrics
}

func calculateTotalVault(accounts map[string]*account.Account) float64 {
	sum := 0.0
	for _, acc := range accounts {
		sum += acc.GetBalance()
	}
	return sum
}

// snapshotGlobal tar rekursivt lås på alle kontoer i deterministisk rekkefølge
// for å verifisere at summen av penger i hvelvet aldri avviker midt under transaksjoner.
func snapshotGlobal(accounts []*account.Account, idx int) float64 {
	if idx == len(accounts) {
		sum := 0.0
		for _, a := range accounts {
			sum += a.BalanceUnsafe()
		}
		return sum
	}

	accounts[idx].Lock()
	defer accounts[idx].Unlock()
	return snapshotGlobal(accounts, idx+1)
}

package simulation_test

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/simulation"
	"github.com/roblar658/concurrent-bank-transaction-go/pkg/transfer"
)

func TestCsvSimulations(t *testing.T) {
	accountsPath := filepath.Join("..", "..", "data", "accounts_100.csv")
	benchmarkPath := filepath.Join("..", "..", "data", "transactions_benchmark_10k.csv")
	adversarialPath := filepath.Join("..", "..", "data", "transactions_adversarial_10k.csv")

	// 1. Benchmark CSV test
	t.Run("Benchmark_10k", func(t *testing.T) {
		accounts, err := simulation.LoadAccounts(accountsPath)
		if err != nil {
			t.Fatalf("Lasting av kontoer feilet: %v", err)
		}
		txs, err := simulation.LoadTransactions(benchmarkPath)
		if err != nil {
			t.Fatalf("Lasting av transaksjoner feilet: %v", err)
		}

		svc := transfer.NewOrderedService()
		engine := simulation.NewEngine(svc)

		metrics := engine.RunSimulation(accounts, txs, 16, true)
		metrics.PrintReport("Go Test: Benchmark 10k CSV", 16)

		if metrics.TotalTransactions() != 10000 {
			t.Errorf("Forventet 10000 transaksjoner, fikk %d", metrics.TotalTransactions())
		}
		if metrics.AuditViolationsCount() > 0 {
			t.Errorf("Audit avvik funnet: %d", metrics.AuditViolationsCount())
		}
		diff := math.Abs(metrics.FinalVault() - metrics.InitialVault())
		if diff > 0.0001 {
			t.Errorf("Bankbalanse ikke bevart, diff: %.4f", diff)
		}
	})

	// 2. Adversarial Deadlock-Trap test
	t.Run("Adversarial_10k", func(t *testing.T) {
		accounts, err := simulation.LoadAccounts(accountsPath)
		if err != nil {
			t.Fatalf("Lasting av kontoer feilet: %v", err)
		}
		txs, err := simulation.LoadTransactions(adversarialPath)
		if err != nil {
			t.Fatalf("Lasting av transaksjoner feilet: %v", err)
		}

		svc := transfer.NewOrderedService()
		engine := simulation.NewEngine(svc)

		metrics := engine.RunSimulation(accounts, txs, 20, true)
		metrics.PrintReport("Go Test: Adversarial 10k CSV", 20)

		if metrics.TotalTransactions() != 10000 {
			t.Errorf("Forventet 10000 transaksjoner, fikk %d", metrics.TotalTransactions())
		}
		if metrics.AuditViolationsCount() > 0 {
			t.Errorf("Audit avvik funnet: %d", metrics.AuditViolationsCount())
		}
		diff := math.Abs(metrics.FinalVault() - metrics.InitialVault())
		if diff > 0.0001 {
			t.Errorf("Bankbalanse ikke bevart, diff: %.4f", diff)
		}
	})
}

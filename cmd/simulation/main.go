package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/simulation"
	"github.com/roblar658/concurrent-bank-transaction-go/pkg/transfer"
)

func main() {
	fmt.Println("================================================================================")
	fmt.Println("     GO BANK SIMULERINGSMOTOR - MASSIV SAMTIDIGHET MED PREDEFINERT CSV-DATA     ")
	fmt.Println("================================================================================")

	accountsPath := "data/accounts_100.csv"
	transactionsPath := "data/transactions_adversarial_10k.csv"
	workers := runtime.NumCPU()
	if workers < 4 {
		workers = 4
	}

	if len(os.Args) >= 2 {
		accountsPath = os.Args[1]
	}
	if len(os.Args) >= 3 {
		transactionsPath = os.Args[2]
	}
	if len(os.Args) >= 4 {
		if n, err := strconv.Atoi(os.Args[3]); err == nil && n > 0 {
			workers = n
		}
	}

	fmt.Printf("Laster inn data fra disk:\n")
	fmt.Printf("  - Kontoer fil     : %s\n", accountsPath)
	fmt.Printf("  - Transaksjoner   : %s\n", transactionsPath)
	fmt.Printf("  - Goroutiner      : %d\n\n", workers)

	loadStart := time.Now()
	accounts, err := simulation.LoadAccounts(accountsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Feil ved lasting av kontoer: %v\n", err)
		os.Exit(1)
	}

	transactions, err := simulation.LoadTransactions(transactionsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Feil ved lasting av transaksjoner: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Data innlastet på %v (%d kontoer, %d transaksjoner).\n",
		time.Since(loadStart), len(accounts), len(transactions))

	transferService := transfer.NewOrderedService()
	engine := simulation.NewEngine(transferService)

	fmt.Println("\nStarter parallell simulering med aktiv sanntidsrevisjon (auditor)...")
	metrics := engine.RunSimulation(accounts, transactions, workers, true)

	metrics.PrintReport(transactionsPath, workers)

	diff := math.Abs(metrics.FinalVault() - metrics.InitialVault())
	if diff > 0.0001 {
		fmt.Fprintf(os.Stderr, "KRITISK FEIL: Totalbankbalansen avviker med %.4f kr!\n", diff)
		os.Exit(1)
	}

	if metrics.AuditViolationsCount() > 0 {
		fmt.Fprintf(os.Stderr, "KRITISK FEIL: %d inkonsistente mellomtilstander oppdaget av auditor!\n",
			metrics.AuditViolationsCount())
		os.Exit(1)
	}

	fmt.Println(">>> GO SIMULERING VELLYKKET: Full konsistens, 0 vranglåser og 0 audit-avvik bekreftet!")
}

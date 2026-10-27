package simulation

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/roblar658/concurrent-bank-transaction-go/pkg/account"
)

// TransactionRecord representerer en overføring innlest fra CSV.
type TransactionRecord struct {
	TxID   string
	FromID string
	ToID   string
	Amount float64
}

// LoadAccounts leser inn kontoer fra en CSV-fil (format: id,owner,balance,is_fair).
func LoadAccounts(filePath string) (map[string]*account.Account, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("kunne ikke åpne konto-CSV: %w", err)
	}
	defer file.Close()

	accounts := make(map[string]*account.Account)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	isHeader := true

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if isHeader {
			isHeader = false
			if strings.Contains(strings.ToLower(line), "balance") {
				continue
			}
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			return nil, fmt.Errorf("ugyldig CSV-linje %d: '%s'", lineNumber, line)
		}

		id := strings.TrimSpace(parts[0])
		owner := strings.TrimSpace(parts[1])
		balance, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		if err != nil {
			return nil, fmt.Errorf("ugyldig saldo på linje %d: %w", lineNumber, err)
		}

		acc, err := account.New(id, owner, balance)
		if err != nil {
			return nil, fmt.Errorf("feil ved opprettelse av konto på linje %d: %w", lineNumber, err)
		}

		if _, exists := accounts[id]; exists {
			return nil, fmt.Errorf("duplikat konto-ID funnet på linje %d: %s", lineNumber, id)
		}
		accounts[id] = acc
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("feil under lesing av CSV: %w", err)
	}

	return accounts, nil
}

// LoadTransactions leser inn overføringstransaksjoner fra en CSV-fil (format: tx_id,from_id,to_id,amount).
func LoadTransactions(filePath string) ([]TransactionRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("kunne ikke åpne transaksjons-CSV: %w", err)
	}
	defer file.Close()

	var transactions []TransactionRecord
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	isHeader := true

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if isHeader {
			isHeader = false
			if strings.Contains(strings.ToLower(line), "from") || strings.Contains(strings.ToLower(line), "to") {
				continue
			}
		}

		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			return nil, fmt.Errorf("ugyldig transaksjonslinje %d: '%s'", lineNumber, line)
		}

		txID := strings.TrimSpace(parts[0])
		fromID := strings.TrimSpace(parts[1])
		toID := strings.TrimSpace(parts[2])
		amount, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil {
			return nil, fmt.Errorf("ugyldig beløp på linje %d: %w", lineNumber, err)
		}

		transactions = append(transactions, TransactionRecord{
			TxID:   txID,
			FromID: fromID,
			ToID:   toID,
			Amount: amount,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("feil under lesing av CSV: %w", err)
	}

	return transactions, nil
}

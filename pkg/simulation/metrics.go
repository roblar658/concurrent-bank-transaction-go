package simulation

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// SimulationMetrics samler inn nøyaktig tidsbruk, gjennomstrømming og latens-persentiler.
type SimulationMetrics struct {
	startTime time.Time
	endTime   time.Time

	totalTransactions      atomic.Uint64
	successfulTransfers    atomic.Uint64
	insufficientFundsCount atomic.Uint64
	selfTransferCount      atomic.Uint64
	invalidAmountCount     atomic.Uint64
	missingAccountCount    atomic.Uint64

	auditChecksCount     atomic.Uint64
	auditViolationsCount atomic.Uint64

	initialVaultBalance float64
	finalVaultBalance   float64

	latenciesMu sync.Mutex
	latenciesUs []int64
}

// NewMetrics oppretter en ny metrikk-samler.
func NewMetrics() *SimulationMetrics {
	return &SimulationMetrics{
		startTime:   time.Now(),
		latenciesUs: make([]int64, 0, 50000),
	}
}

func (m *SimulationMetrics) RecordSuccess(d time.Duration) {
	m.totalTransactions.Add(1)
	m.successfulTransfers.Add(1)
	us := d.Microseconds()

	m.latenciesMu.Lock()
	m.latenciesUs = append(m.latenciesUs, us)
	m.latenciesMu.Unlock()
}

func (m *SimulationMetrics) RecordInsufficientFunds(d time.Duration) {
	m.totalTransactions.Add(1)
	m.insufficientFundsCount.Add(1)
	us := d.Microseconds()

	m.latenciesMu.Lock()
	m.latenciesUs = append(m.latenciesUs, us)
	m.latenciesMu.Unlock()
}

func (m *SimulationMetrics) RecordSelfTransfer() {
	m.totalTransactions.Add(1)
	m.selfTransferCount.Add(1)
}

func (m *SimulationMetrics) RecordInvalidAmount() {
	m.totalTransactions.Add(1)
	m.invalidAmountCount.Add(1)
}

func (m *SimulationMetrics) RecordMissingAccount() {
	m.totalTransactions.Add(1)
	m.missingAccountCount.Add(1)
}

func (m *SimulationMetrics) RecordAudit(consistent bool) {
	m.auditChecksCount.Add(1)
	if !consistent {
		m.auditViolationsCount.Add(1)
	}
}

func (m *SimulationMetrics) Complete(initialVault, finalVault float64) {
	m.endTime = time.Now()
	m.initialVaultBalance = initialVault
	m.finalVaultBalance = finalVault
}

type LatencyStats struct {
	Min int64
	Avg float64
	P50 int64
	P90 int64
	P95 int64
	P99 int64
	Max int64
}

func (m *SimulationMetrics) CalculateLatencies() LatencyStats {
	m.latenciesMu.Lock()
	defer m.latenciesMu.Unlock()

	n := len(m.latenciesUs)
	if n == 0 {
		return LatencyStats{}
	}

	sorted := make([]int64, n)
	copy(sorted, m.latenciesUs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var sum int64
	for _, v := range sorted {
		sum += v
	}

	return LatencyStats{
		Min: sorted[0],
		Avg: float64(sum) / float64(n),
		P50: sorted[int(float64(n)*0.50)],
		P90: sorted[int(float64(n)*0.90)],
		P95: sorted[int(float64(n)*0.95)],
		P99: sorted[int(float64(n)*0.99)],
		Max: sorted[n-1],
	}
}

func (m *SimulationMetrics) PrintReport(title string, workerGoroutines int) {
	dur := m.endTime.Sub(m.startTime)
	durSec := dur.Seconds()
	total := m.totalTransactions.Load()
	tps := 0.0
	if durSec > 0 {
		tps = float64(total) / durSec
	}
	lat := m.CalculateLatencies()
	succ := m.successfulTransfers.Load()
	succPct := 0.0
	if total > 0 {
		succPct = (float64(succ) / float64(total)) * 100.0
	}

	fmt.Println("\n================================================================================")
	fmt.Printf("  GO SIMULERINGSRAPPORT: %s\n", title)
	fmt.Println("================================================================================")
	fmt.Printf("  Arbeids-goroutiner             : %d goroutines\n", workerGoroutines)
	fmt.Printf("  Total varighet                 : %.3f sekunder (%.1f ms)\n", durSec, float64(dur.Milliseconds()))
	fmt.Printf("  Totalt behandlet               : %d transaksjoner\n", total)
	fmt.Printf("  -> Vellykkede overføringer     : %d (%.2f%%)\n", succ, succPct)
	fmt.Printf("  -> Avvist (manglende dekning)  : %d\n", m.insufficientFundsCount.Load())
	fmt.Printf("  -> Avvist (egen-overføring)    : %d\n", m.selfTransferCount.Load())
	fmt.Printf("  -> Avvist (ugyldig beløp)      : %d\n", m.invalidAmountCount.Load())
	fmt.Printf("  -> Feil (ukjent konto)         : %d\n", m.missingAccountCount.Load())
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("  Gjennomstrømming (Throughput)  : %.1f transaksjoner/sekund (TPS)\n", tps)
	fmt.Printf("  Latens (Min / Gj.snitt / Maks) : %d µs / %.2f µs / %d µs\n", lat.Min, lat.Avg, lat.Max)
	fmt.Printf("  Latens-persentiler (p50 / p95) : %d µs / %d µs\n", lat.P50, lat.P95)
	fmt.Printf("  Latens-persentil   (p99)       : %d µs (%.3f ms)\n", lat.P99, float64(lat.P99)/1000.0)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("  Start-saldo bankhvelv          : %.2f kr\n", m.initialVaultBalance)
	fmt.Printf("  Slutt-saldo bankhvelv          : %.2f kr\n", m.finalVaultBalance)
	fmt.Printf("  Differanse i hvelv             : %.4f kr\n", m.finalVaultBalance-m.initialVaultBalance)
	fmt.Printf("  Kontinuerlige audit-sjekker    : %d utført, %d avvik funnet\n",
		m.auditChecksCount.Load(), m.auditViolationsCount.Load())
	fmt.Println("================================================================================")
}

func (m *SimulationMetrics) TotalTransactions() uint64      { return m.totalTransactions.Load() }
func (m *SimulationMetrics) SuccessfulTransfers() uint64    { return m.successfulTransfers.Load() }
func (m *SimulationMetrics) AuditViolationsCount() uint64   { return m.auditViolationsCount.Load() }
func (m *SimulationMetrics) InitialVault() float64          { return m.initialVaultBalance }
func (m *SimulationMetrics) FinalVault() float64            { return m.finalVaultBalance }

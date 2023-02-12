package metrics

type Metrics struct {
	FailureCount int
}

func (m *Metrics) ReportFailure() {
	m.FailureCount++
}

func (m *Metrics) GetFailure() int {
	return m.FailureCount
}

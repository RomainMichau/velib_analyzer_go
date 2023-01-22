package main

type Metrics struct {
	failureCount int
}

func (m *Metrics) reportFailure() {
	m.failureCount++
}

func (m *Metrics) getFailure() int {
	return m.failureCount
}

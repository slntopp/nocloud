package billing

import "sync"

var m = &sync.Mutex{}
var transactionGates = map[string]*sync.Mutex{}

func closeTransactionGate(account string) {
	m.Lock()
	gate, ok := transactionGates[account]
	if !ok {
		gate = &sync.Mutex{}
		transactionGates[account] = gate
	}
	m.Unlock()
	gate.Lock()
}

func openTransactionGate(account string) {
	m.Lock()
	gate, ok := transactionGates[account]
	if !ok {
		gate = &sync.Mutex{}
		transactionGates[account] = gate
	}
	m.Unlock()
	gate.Unlock()
}

package sdk

import (
	"testing"
)

func TestNewTransaction(t *testing.T) {

	transaction := NewTransaction()

	if transaction.status != statusUnknown {
		t.Error("new transaction should have status unknown")
	}

	if transaction.state != stateOk {
		t.Error("new transaction should have state ok")
	}

	if transaction.created != transaction.updated {
		t.Error("new transaction created and updated times should be the same")
	}

	if transaction.message != "" {
		t.Error("new transaction should not have a message")
	}

	tr, found := transactionCache.Get(transaction.id)
	if !found {
		t.Error("new transaction not added to the cache")
	}
	if tr != transaction {
		t.Error("cached value does not equal returned value")
	}
}


func TestNewTransaction2(t *testing.T) {
	t1 := NewTransaction()
	t2 := NewTransaction()

	if t1.id == t2.id {
		t.Error("two new transactions have the same id")
	}

	if t1.status != t2.status {
		t.Error("new transaction statuses should be the same")
	}

	if t1.state != t2.state {
		t.Error("new transaction states should be the same")
	}

	if t1.message != t2.message {
		t.Error("new transaction messages should be the same")
	}

	tr, found := transactionCache.Get(t1.id)
	if !found {
		t.Error("new transaction not added to the cache")
	}
	if tr != t1 {
		t.Error("cached value does not equal returned value")
	}

	tr, found = transactionCache.Get(t2.id)
	if !found {
		t.Error("new transaction not added to the cache")
	}
	if tr != t2 {
		t.Error("cached value does not equal returned value")
	}
}

func TestGetTransaction(t *testing.T) {
	tr := NewTransaction()

	transaction := GetTransaction(tr.id)
	if transaction != tr {
		t.Error("did not find transaction in the cache")
	}
}

func TestGetTransaction2(t *testing.T) {
	transaction := GetTransaction("123")
	if transaction != nil {
		t.Error("found transaction in the cache that should not be there")
	}
}

func TestTransaction_setStateOk(t *testing.T) {
	tr := NewTransaction()

	tr.state = stateError
	if tr.state != stateError {
		t.Error("transaction should be in the error state")
	}

	tr.setStateOk()
	if tr.state != stateOk {
		t.Error("failed to set transaction state to Ok")
	}
}

func TestTransaction_setStateOkCached(t *testing.T) {
	tr := NewTransaction()
	cached := GetTransaction(tr.id)

	cached.state = stateError
	if tr.state != stateError {
		t.Error("transaction should be in the error state")
	}

	cached.setStateOk()
	if cached.state != stateOk {
		t.Error("failed to set transaction state to Ok")
	}
	if tr.state != stateOk {
		t.Error("failed to set transaction state to Ok via cached ref")
	}
}

func TestTransaction_setStateErr(t *testing.T) {
	tr := NewTransaction()

	tr.state = stateOk
	if tr.state != stateOk {
		t.Error("transaction should be in the ok state")
	}

	tr.setStateError()
	if tr.state != stateError {
		t.Error("failed to set transaction state to error")
	}
}

func TestTransaction_setStateErrCached(t *testing.T) {
	tr := NewTransaction()
	cached := GetTransaction(tr.id)

	cached.state = stateOk
	if tr.state != stateOk {
		t.Error("transaction should be in the ok state")
	}

	cached.setStateError()
	if cached.state != stateError {
		t.Error("failed to set transaction state to error")
	}
	if tr.state != stateError {
		t.Error("failed to set transaction state to error via cached ref")
	}
}

func TestTransaction_setStatusUnknown(t *testing.T) {
	tr := NewTransaction()

	tr.status = statusDone
	if tr.status != statusDone {
		t.Error("transaction should have the done status")
	}

	tr.setStatusUnknown()
	if tr.status != statusUnknown {
		t.Error("failed to set transaction status to unknown")
	}
}

func TestTransaction_setStatusUnknownCached(t *testing.T) {
	tr := NewTransaction()
	cached := GetTransaction(tr.id)

	cached.status = statusDone
	if tr.status != statusDone {
		t.Error("transaction should have the done status")
	}

	cached.setStatusUnknown()
	if cached.status != statusUnknown {
		t.Error("failed to set transaction status to unknown")
	}
	if tr.status != statusUnknown {
		t.Error("failed to set transaction status to unknown via cached ref")
	}
}

func TestTransaction_setStatusPending(t *testing.T) {
	tr := NewTransaction()

	tr.status = statusUnknown
	if tr.status != statusUnknown {
		t.Error("transaction should have the unknown status")
	}

	tr.setStatusPending()
	if tr.status != statusPending {
		t.Error("failed to set transaction status to pending")
	}
}

func TestTransaction_setStatusPendingCached(t *testing.T) {
	tr := NewTransaction()
	cached := GetTransaction(tr.id)

	cached.status = statusUnknown
	if tr.status != statusUnknown {
		t.Error("transaction should have the unknown status")
	}

	cached.setStatusPending()
	if cached.status != statusPending {
		t.Error("failed to set transaction status to pending")
	}
	if tr.status != statusPending {
		t.Error("failed to set transaction status to pending via cached ref")
	}
}

func TestTransaction_setStatusWriting(t *testing.T) {
	tr := NewTransaction()

	tr.status = statusUnknown
	if tr.status != statusUnknown {
		t.Error("transaction should have the unknown status")
	}

	tr.setStatusWriting()
	if tr.status != statusWriting {
		t.Error("failed to set transaction status to writing")
	}
}

func TestTransaction_setStatusWritingCached(t *testing.T) {
	tr := NewTransaction()
	cached := GetTransaction(tr.id)

	cached.status = statusUnknown
	if tr.status != statusUnknown {
		t.Error("transaction should have the unknown status")
	}

	cached.setStatusWriting()
	if cached.status != statusWriting {
		t.Error("failed to set transaction status to writing")
	}
	if tr.status != statusWriting {
		t.Error("failed to set transaction status to writing via cached ref")
	}
}

func TestTransaction_setStatusDone(t *testing.T) {
	tr := NewTransaction()

	tr.status = statusUnknown
	if tr.status != statusUnknown {
		t.Error("transaction should have the unknown status")
	}

	tr.setStatusDone()
	if tr.status != statusDone {
		t.Error("failed to set transaction status to done")
	}
}

func TestTransaction_setStatusDoneCached(t *testing.T) {
	tr := NewTransaction()
	cached := GetTransaction(tr.id)

	cached.status = statusUnknown
	if tr.status != statusUnknown {
		t.Error("transaction should have the unknown status")
	}

	cached.setStatusDone()
	if cached.status != statusDone {
		t.Error("failed to set transaction status to done")
	}
	if tr.status != statusDone {
		t.Error("failed to set transaction status to done via cached ref")
	}
}

func TestTransaction_encode(t *testing.T) {
	tr := NewTransaction()
	encoded := tr.encode()

	if encoded.Status != tr.status {
		t.Error("encoded status does not match transaction status")
	}

	if encoded.State != tr.state {
		t.Error("encoded state does not match transaction state")
	}

	if encoded.Created != tr.created {
		t.Error("encoded created does not match transaction created")
	}

	if encoded.Updated != tr.updated {
		t.Error("encoded updated does not match transaction updated")
	}

	if encoded.Message != tr.message {
		t.Error("encoded message does not match transaction message")
	}
}
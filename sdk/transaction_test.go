package sdk

import (
	"testing"
)

// Helper method to call NewTransaction and fail the test if it does not
// succeed. Returns the new transaction.
func NewTransactionHelper(t *testing.T) *Transaction {
	transaction, err := NewTransaction()
	if err != nil {
		t.Errorf("NewTransaction should succeed. err %v", err)
	}
	return transaction
}

// Helper method to call GetTransaction and fail the test if it does not
// succeed. Returns the transaction retrieved.
func GetTransactionHelper(t *testing.T, id string) *Transaction {
	transaction, err := GetTransaction(id)
	if err != nil {
		t.Errorf("GetTransaction should succeed. err %v", err)
	}
	return transaction
}

func TestNewTransaction(t *testing.T) {
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)

	transaction := NewTransactionHelper(t)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)

	t1 := NewTransactionHelper(t)
	t2 := NewTransactionHelper(t)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)

	cached := GetTransactionHelper(t, tr.id)
	if cached != tr {
		t.Error("did not find transaction in the cache")
	}
}

func TestGetTransaction2(t *testing.T) {
	transaction, err := GetTransaction("123")
	if transaction != nil {
		t.Error("found transaction in the cache that should not be there")
	}
	if err != nil {
		t.Errorf(
			"Trying to get a transaction that is not in the cache is not an error. err %v",
			err)
	}
}

func TestTransaction_setStateOk(t *testing.T) {
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)
	cached := GetTransactionHelper(t, tr.id)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)
	cached := GetTransactionHelper(t, tr.id)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)
	cached := GetTransactionHelper(t, tr.id)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)
	cached := GetTransactionHelper(t, tr.id)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)

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
	tr := NewTransactionHelper(t)
	cached := GetTransactionHelper(t, tr.id)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)
	cached := GetTransactionHelper(t, tr.id)

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
	// Setup the transaction cache so NewTransaction can succeed.
	// Ignore any error since a previous test may have set it up.
	SetupTransactionCache(600)
	tr := NewTransactionHelper(t)
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

func TestTransactionCache(t *testing.T) {
	// Setup the transaction cache without error checking so we know it is initialized.
	SetupTransactionCache(600)

	// Invalidate the transaction cache. Error should be nil.
	err := InvalidateTransactionCache()
	if err != nil {
		t.Errorf("InvalidateTransactionCache should not fail. err: %v", err)
	}

	// Invalidate the transaction cache again. Error should not be nil.
	err = InvalidateTransactionCache()
	if err == nil {
		t.Errorf("InvalidateTransactionCache should fail. err: %v", err)
	}

	// Call NewTransaction. Should fail without the transaction cache.
	_, err = NewTransaction()
	if err == nil {
		t.Errorf("NewTransaction should fail. err: %v", err)
	}

	// Call GetTransaction. Should fail without the transaction cache.
	tr, err := GetTransaction("zzz")
	if err == nil {
		t.Errorf("GetTransaction should fail. err: %v", err)
	}
	if tr != nil {
		t.Errorf("GetTransaction should not return a transaction. tr: %v", tr)
	}

	// Call SetupTransactionCache. Should not error.
	err = SetupTransactionCache(600)
	if err != nil {
		t.Errorf("SetupTransactionCache should not fail. err: %v", err)
	}

	// Call SetupTransactionCache again. Should error.
	err = SetupTransactionCache(600)
	if err == nil {
		t.Errorf("SetupTransactionCache should fail. err: %v", err)
	}
}

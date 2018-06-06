package sdk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewTransaction tests creating a new transaction and getting
// it back out from the cache.
func TestNewTransaction(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)

	transaction := newTransaction()

	assert.Equal(t, statusUnknown, transaction.status)
	assert.Equal(t, stateOk, transaction.state)
	assert.Equal(t, transaction.created, transaction.updated)
	assert.Equal(t, "", transaction.message)

	tr, found := transactionCache.Get(transaction.id)
	assert.True(t, found, "new transaction not added to the cache")
	assert.Equal(t, transaction, tr, "cache value does not equal returned value")
}

// TestNewTransaction2 tests creating new transactions and getting
// them back out from the cache.
func TestNewTransaction2(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)

	t1 := newTransaction()
	t2 := newTransaction()

	assert.NotEqual(t, t1.id, t2.id, "two transactions should not have the same id")

	assert.Equal(t, t1.status, t2.status)
	assert.Equal(t, t1.state, t2.state)
	assert.Equal(t, t1.message, t2.message)

	tr, found := transactionCache.Get(t1.id)
	assert.True(t, found, "new transaction not added to the cache")
	assert.Equal(t, t1, tr, "cache value does not equal returned value")

	tr, found = transactionCache.Get(t2.id)
	assert.True(t, found, "new transaction not added to the cache")
	assert.Equal(t, t2, tr, "cache value does not equal returned value")
}

// TestGetTransaction tests getting a transaction from the cache.
func TestGetTransaction(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	cached := getTransaction(tr.id)
	assert.Equal(t, tr, cached)
}

// TestGetTransaction2 tests getting a transaction from the cache
// that does not exist in the cache.
func TestGetTransaction2(t *testing.T) {
	transaction := getTransaction("123")
	assert.Nil(t, transaction)
}

// TestTransaction_setStateOk tests setting the state of a transaction to OK.
func TestTransaction_setStateOk(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.state = stateError
	assert.Equal(t, stateError, tr.state)

	tr.setStateOk()
	assert.Equal(t, stateOk, tr.state)
}

// TestTransaction_setStateOkCached tests setting the state of a cached
// transaction to OK.
func TestTransaction_setStateOkCached(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.state = stateError
	assert.Equal(t, stateError, cached.state)
	assert.Equal(t, stateError, tr.state)

	cached.setStateOk()
	assert.Equal(t, stateOk, cached.state)
	assert.Equal(t, stateOk, tr.state)
}

// TestTransaction_setStateErr tests setting the state of a transaction to Error.
func TestTransaction_setStateErr(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.state = stateOk
	assert.Equal(t, stateOk, tr.state)

	tr.setStateError()
	assert.Equal(t, stateError, tr.state)
}

// TestTransaction_setStateErrCached tests setting the state of a cached
// transaction to Error.
func TestTransaction_setStateErrCached(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.state = stateOk
	assert.Equal(t, stateOk, cached.state)
	assert.Equal(t, stateOk, tr.state)

	cached.setStateError()
	assert.Equal(t, stateError, cached.state)
	assert.Equal(t, stateError, tr.state)
}

// TestTransaction_setStatusUnknown tests setting the status of a transaction to Unknown.
func TestTransaction_setStatusUnknown(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.status = statusDone
	assert.Equal(t, statusDone, tr.status)

	tr.setStatusUnknown()
	assert.Equal(t, statusUnknown, tr.status)
}

// TestTransaction_setStatusUnknownCached tests setting the status of a cached
// transaction to Unknown.
func TestTransaction_setStatusUnknownCached(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.status = statusDone
	assert.Equal(t, statusDone, cached.status)
	assert.Equal(t, statusDone, tr.status)

	cached.setStatusUnknown()
	assert.Equal(t, statusUnknown, cached.status)
	assert.Equal(t, statusUnknown, tr.status)
}

// TestTransaction_setStatusPending tests setting the status of a transaction to Pending.
func TestTransaction_setStatusPending(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.status = statusUnknown
	assert.Equal(t, statusUnknown, tr.status)

	tr.setStatusPending()
	assert.Equal(t, statusPending, tr.status)
}

// TestTransaction_setStatusPendingCached tests setting the status of a cached
// transaction to Pending.
func TestTransaction_setStatusPendingCached(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.status = statusUnknown
	assert.Equal(t, statusUnknown, cached.status)
	assert.Equal(t, statusUnknown, tr.status)

	cached.setStatusPending()
	assert.Equal(t, statusPending, cached.status)
	assert.Equal(t, statusPending, tr.status)
}

// TestTransaction_setStatusWriting tests setting the status of a transaction to Writing.
func TestTransaction_setStatusWriting(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.status = statusUnknown
	assert.Equal(t, statusUnknown, tr.status)

	tr.setStatusWriting()
	assert.Equal(t, statusWriting, tr.status)
}

// TestTransaction_setStatusWritingCached tests setting the status of a cached
// transaction to Writing.
func TestTransaction_setStatusWritingCached(t *testing.T) {
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.status = statusUnknown
	assert.Equal(t, statusUnknown, cached.status)
	assert.Equal(t, statusUnknown, tr.status)

	cached.setStatusWriting()
	assert.Equal(t, statusWriting, cached.status)
	assert.Equal(t, statusWriting, tr.status)
}

// TestTransaction_setStatusDone tests setting the status of a transaction to Done.
func TestTransaction_setStatusDone(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.status = statusUnknown
	assert.Equal(t, statusUnknown, tr.status)

	tr.setStatusDone()
	assert.Equal(t, statusDone, tr.status)
}

// TestTransaction_setStatusDoneCached tests setting the status of a cached
// transaction to Done.
func TestTransaction_setStatusDoneCached(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.status = statusUnknown
	assert.Equal(t, statusUnknown, cached.status)
	assert.Equal(t, statusUnknown, tr.status)

	cached.setStatusDone()
	assert.Equal(t, statusDone, cached.status)
	assert.Equal(t, statusDone, tr.status)
}

// TestTransaction_encode tests encoding an SDK transaction into the
// gRPC transaction struct.
func TestTransaction_encode(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()
	encoded := tr.encode()

	assert.Equal(t, tr.status, encoded.Status)
	assert.Equal(t, tr.state, encoded.State)
	assert.Equal(t, tr.created, encoded.Created)
	assert.Equal(t, tr.updated, encoded.Updated)
	assert.Equal(t, tr.message, encoded.Message)
}

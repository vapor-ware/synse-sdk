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

	assert.Equal(t, statusPending, transaction.status)
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


// TestTransaction_setStatusError tests setting the status of a transaction to Error.
func TestTransaction_setStatusError(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.status = statusDone
	assert.Equal(t, statusDone, tr.status)

	tr.setStatusError()
	assert.Equal(t, statusError, tr.status)
}

// TestTransaction_setStatusErrorCached tests setting the status of a cached
// transaction to Error.
func TestTransaction_setStatusErrorCached(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.status = statusDone
	assert.Equal(t, statusDone, cached.status)
	assert.Equal(t, statusDone, tr.status)

	cached.setStatusError()
	assert.Equal(t, statusError, cached.status)
	assert.Equal(t, statusError, tr.status)
}

// TestTransaction_setStatusPending tests setting the status of a transaction to Pending.
func TestTransaction_setStatusPending(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.status = statusDone
	assert.Equal(t, statusDone, tr.status)

	tr.setStatusPending()
	assert.Equal(t, statusPending, tr.status)
}

// TestTransaction_setStatusPendingCached tests setting the status of a cached
// transaction to Pending.
func TestTransaction_setStatusPendingCached(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.status = statusDone
	assert.Equal(t, statusDone, cached.status)
	assert.Equal(t, statusDone, tr.status)

	cached.setStatusPending()
	assert.Equal(t, statusPending, cached.status)
	assert.Equal(t, statusPending, tr.status)
}

// TestTransaction_setStatusWriting tests setting the status of a transaction to Writing.
func TestTransaction_setStatusWriting(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.status = statusDone
	assert.Equal(t, statusDone, tr.status)

	tr.setStatusWriting()
	assert.Equal(t, statusWriting, tr.status)
}

// TestTransaction_setStatusWritingCached tests setting the status of a cached
// transaction to Writing.
func TestTransaction_setStatusWritingCached(t *testing.T) {
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.status = statusDone
	assert.Equal(t, statusDone, cached.status)
	assert.Equal(t, statusDone, tr.status)

	cached.setStatusWriting()
	assert.Equal(t, statusWriting, cached.status)
	assert.Equal(t, statusWriting, tr.status)
}

// TestTransaction_setStatusDone tests setting the status of a transaction to Done.
func TestTransaction_setStatusDone(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()

	tr.status = statusPending
	assert.Equal(t, statusPending, tr.status)

	tr.setStatusDone()
	assert.Equal(t, statusDone, tr.status)
}

// TestTransaction_setStatusDoneCached tests setting the status of a cached
// transaction to Done.
func TestTransaction_setStatusDoneCached(t *testing.T) {
	setupTransactionCache(time.Duration(600) * time.Second)
	tr := newTransaction()
	cached := getTransaction(tr.id)

	cached.status = statusPending
	assert.Equal(t, statusPending, cached.status)
	assert.Equal(t, statusPending, tr.status)

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
	assert.Equal(t, tr.created, encoded.Created)
	assert.Equal(t, tr.updated, encoded.Updated)
	assert.Equal(t, tr.message, encoded.Message)
}

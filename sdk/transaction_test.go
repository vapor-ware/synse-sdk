// Synse SDK
// Copyright (c) 2017-2022 Vapor IO
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var defaultTimeout = 5 * time.Second

// TestNewTransaction tests creating a new transaction and getting
// it back out from the cache.
func TestNewTransaction(t *testing.T) {
	transaction := newTransaction(defaultTimeout, "")

	assert.Equal(t, statusPending, transaction.status)
	assert.Equal(t, transaction.created, transaction.updated)
	assert.Equal(t, "", transaction.message)
}

// TestNewTransaction2 tests creating new transactions and getting
// them back out from the cache.
func TestNewTransaction2(t *testing.T) {
	t1 := newTransaction(defaultTimeout, "")
	t2 := newTransaction(defaultTimeout, "")

	assert.NotEqual(t, t1.id, t2.id, "two transactions should not have the same id")

	assert.Equal(t, t1.status, t2.status)
	assert.Equal(t, t1.message, t2.message)
}

// TestNewTransaction3 tests creating a new transaction with a custom ID.
func TestNewTransaction3(t *testing.T) {
	txn := newTransaction(defaultTimeout, "abc123")

	assert.Equal(t, statusPending, txn.status)
	assert.Equal(t, txn.created, txn.updated)
	assert.Equal(t, "", txn.message)
	assert.Equal(t, defaultTimeout, txn.timeout)
	assert.Equal(t, "abc123", txn.id)
}

// TestNewTransaction4 tests creating multiple new transactions with the same custom ID.
// It is not the onus of the constructor to ensure IDs are unique -- that is left to
// whatever puts the transaction in the cache.
func TestNewTransaction4(t *testing.T) {
	t1 := newTransaction(defaultTimeout, "abc123")
	t2 := newTransaction(defaultTimeout, "abc123")

	assert.Equal(t, statusPending, t1.status)
	assert.Equal(t, t1.created, t1.updated)
	assert.Equal(t, "", t1.message)
	assert.Equal(t, defaultTimeout, t1.timeout)
	assert.Equal(t, "abc123", t1.id)

	assert.Equal(t, t1.status, t2.status)
	assert.Equal(t, t1.message, t2.message)
	assert.Equal(t, t1.timeout, t2.timeout)
	assert.Equal(t, t1.id, t2.id)
}

// TestTransaction_setStatusError tests setting the status of a transaction to Error.
func TestTransaction_setStatusError(t *testing.T) {
	tr := newTransaction(defaultTimeout, "")

	tr.status = statusDone
	assert.Equal(t, statusDone, tr.status)

	tr.setStatusError()
	assert.Equal(t, statusError, tr.status)
}

// TestTransaction_setStatusPending tests setting the status of a transaction to Pending.
func TestTransaction_setStatusPending(t *testing.T) {
	tr := newTransaction(defaultTimeout, "")

	tr.status = statusDone
	assert.Equal(t, statusDone, tr.status)

	tr.setStatusPending()
	assert.Equal(t, statusPending, tr.status)
}

// TestTransaction_setStatusWriting tests setting the status of a transaction to Writing.
func TestTransaction_setStatusWriting(t *testing.T) {
	tr := newTransaction(defaultTimeout, "")

	tr.status = statusDone
	assert.Equal(t, statusDone, tr.status)

	tr.setStatusWriting()
	assert.Equal(t, statusWriting, tr.status)
}

// TestTransaction_setStatusDone tests setting the status of a transaction to Done.
func TestTransaction_setStatusDone(t *testing.T) {
	tr := newTransaction(defaultTimeout, "")

	tr.status = statusPending
	assert.Equal(t, statusPending, tr.status)

	tr.setStatusDone()
	assert.Equal(t, statusDone, tr.status)
}

// TestTransaction_encode tests encoding an SDK transaction into the
// gRPC transaction struct.
func TestTransaction_encode(t *testing.T) {
	tr := newTransaction(defaultTimeout, "")
	encoded := tr.encode()

	assert.Equal(t, tr.status, encoded.Status)
	assert.Equal(t, tr.created, encoded.Created)
	assert.Equal(t, tr.updated, encoded.Updated)
	assert.Equal(t, tr.message, encoded.Message)
}

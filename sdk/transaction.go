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
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/v2/sdk/utils"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)

const (
	statusPending = synse.WriteStatus_PENDING
	statusWriting = synse.WriteStatus_WRITING
	statusDone    = synse.WriteStatus_DONE
	statusError   = synse.WriteStatus_ERROR
)

// transaction represents an asynchronous write transaction for the Plugin. It
// tracks the state and status of that transaction over its lifetime.
type transaction struct {
	id      string
	status  synse.WriteStatus
	created string
	updated string
	message string
	timeout time.Duration
	context *synse.V3WriteData
	done    chan struct{}
}

// newTransaction creates a new transaction instance.
func newTransaction(timeout time.Duration, customID string) *transaction {
	now := utils.GetCurrentTime()

	var transactionID = customID
	if transactionID == "" {
		transactionID = uuid.New().String()
	}

	return &transaction{
		id:      transactionID,
		status:  statusPending,
		created: now,
		updated: now,
		timeout: timeout,
		message: "",
		done:    make(chan struct{}),
	}
}

// wait waits until the transaction reaches a terminal state (ok, error).
func (t *transaction) wait() {
	log.WithField("id", t.id).Debug("[transaction] waiting for transaction to complete")
	<-t.done
	log.WithField("id", t.id).Debug("[transaction] transaction completed")
}

// encode translates the transaction to a corresponding gRPC V3TransactionStatus.
func (t *transaction) encode() *synse.V3TransactionStatus {
	return &synse.V3TransactionStatus{
		Id:      t.id,
		Created: t.created,
		Updated: t.updated,
		Message: t.message,
		Timeout: t.timeout.String(),
		Status:  t.status,
		Context: t.context,
	}
}

// setStatusPending sets the transaction status to 'pending'.
func (t *transaction) setStatusPending() {
	log.WithField("id", t.id).Debug("[transaction] transaction status set to PENDING")
	t.updated = utils.GetCurrentTime()
	t.status = statusPending
}

// setStatusWriting sets the transaction status to 'writing'.
func (t *transaction) setStatusWriting() {
	log.WithField("id", t.id).Debug("[transaction] transaction status set to WRITING")
	t.updated = utils.GetCurrentTime()
	t.status = statusWriting
}

// setStatusDone sets the transaction status to 'done'.
func (t *transaction) setStatusDone() {
	log.WithField("id", t.id).Debug("[transaction] transaction status set to DONE")
	t.updated = utils.GetCurrentTime()
	t.status = statusDone

	// This is a terminal state, so close the done channel to unblock
	// anything waiting on the transaction to complete.
	close(t.done)
}

// setStatusError sets the transaction status to 'error'.
func (t *transaction) setStatusError() {
	log.WithField("id", t.id).Debug("[transaction] transaction status set to ERROR")
	t.updated = utils.GetCurrentTime()
	t.status = statusError

	// This is a terminal state, so close the done channel to unblock
	// anything waiting on the transaction to complete.
	close(t.done)
}

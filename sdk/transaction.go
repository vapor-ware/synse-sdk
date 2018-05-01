package sdk

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/xid"
	"github.com/vapor-ware/synse-server-grpc/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vapor-ware/synse-sdk/sdk/logger"
)

const (
	statusUnknown = synse.WriteResponse_UNKNOWN
	statusPending = synse.WriteResponse_PENDING
	statusWriting = synse.WriteResponse_WRITING
	statusDone    = synse.WriteResponse_DONE

	stateOk    = synse.WriteResponse_OK
	stateError = synse.WriteResponse_ERROR
)

// transactionCache is a cache with a configurable default expiration time that
// is used to track the asynchronous write transactions as they are processed.
var transactionCache *cache.Cache

// setupTransactionCache creates the transaction cache with the TTL in seconds.
// This needs to be called in order to have transactions.
// If this is called more than once, the cache is reinitialized and an error is
// returned. The caller may choose to ignore the error.
func setupTransactionCache(ttl time.Duration) (err error) {
	// FIXME -- if the cache is not nil, we will return an error, but
	// in that same case, we will still create a new cache.. this seems
	// weird. shouldn't we just return the error in this block and return
	// nil at the end of the func?
	if transactionCache != nil {
		err = status.Errorf(codes.AlreadyExists,
			"transactionCache already initialized.")
	}

	transactionCache = cache.New(
		ttl,
		ttl*2,
	)
	return err
}

// newTransaction creates a new transaction instance. Upon creation, the
// transaction is given a unique ID and is added to the transaction cache.
// This call will fail if the transaction cache is not initialized.
func newTransaction() (*transaction, error) {
	if transactionCache == nil {
		return nil, status.Errorf(codes.FailedPrecondition,
			"transactionCache not initialized. Call setupTransactionCache.")
	}

	id := xid.New().String()
	now := GetCurrentTime()
	t := transaction{
		id:      id,
		status:  statusUnknown,
		state:   stateOk,
		created: now,
		updated: now,
		message: "",
	}
	transactionCache.Set(id, &t, cache.DefaultExpiration)
	return &t, nil
}

// getTransaction looks up the given transaction ID in the cache. If it exists,
// that transaction is returned; otherwise nil is returned.
// This call will fail if the transaction cache is not initialized.
func getTransaction(id string) (*transaction, error) {
	if transactionCache == nil {
		return nil, status.Errorf(codes.FailedPrecondition,
			"transactionCache not initialized. Call setupTransactionCache.")
	}

	t, found := transactionCache.Get(id)
	if found {
		return t.(*transaction), nil
	}
	logger.Info("transaction %v not found", id)
	return nil, nil
}

// transaction represents an asynchronous write transaction for the Plugin. It
// tracks the state and status of that transaction over its lifetime.
type transaction struct {
	id      string
	status  synse.WriteResponse_WriteStatus
	state   synse.WriteResponse_WriteState
	created string
	updated string
	message string
}

// encode translates the transaction to a corresponding gRPC WriteResponse.
func (t *transaction) encode() *synse.WriteResponse {
	return &synse.WriteResponse{
		Status:  t.status,
		State:   t.state,
		Created: t.created,
		Updated: t.updated,
		Message: t.message,
	}
}

// setStateOk sets the transaction to be in the 'ok' state. Since a pointer
// to this struct is stored in the cache, and update here should update the
// in-memory cache as well.
func (t *transaction) setStateOk() {
	logger.Debugf("transaction %v: setting state to 'ok'", t.id)
	t.updated = GetCurrentTime()
	t.state = stateOk
}

// setStateError sets the transaction to be in the 'error' state. Since a
// pointer to this struct is stored in the cache, and update here should
// update the in-memory cache as well.
func (t *transaction) setStateError() {
	logger.Debugf("transaction %v: setting state to 'error'", t.id)
	t.updated = GetCurrentTime()
	t.state = stateError
}

// setStatusUnknown sets the transaction status to 'unknown'. Since a
// pointer to this struct is stored in the cache, and update here should
// update the in-memory cache as well.
func (t *transaction) setStatusUnknown() {
	logger.Debugf("transaction %v: setting status to 'unknown'", t.id)
	t.updated = GetCurrentTime()
	t.status = statusUnknown
}

// setStatusPending sets the transaction status to 'pending'. Since a
// pointer to this struct is stored in the cache, and update here should
// update the in-memory cache as well.
func (t *transaction) setStatusPending() {
	logger.Debugf("transaction %v: setting status to 'pending'", t.id)
	t.updated = GetCurrentTime()
	t.status = statusPending
}

// setStatusWriting sets the transaction status to 'writing'. Since a
// pointer to this struct is stored in the cache, and update here should
// update the in-memory cache as well.
func (t *transaction) setStatusWriting() {
	logger.Debugf("transaction %v: setting status to 'writing'", t.id)
	t.updated = GetCurrentTime()
	t.status = statusWriting
}

// setStatusDone sets the transaction status to 'done'. Since a pointer
// to this struct is stored in the cache, and update here should update
// the in-memory cache as well.
func (t *transaction) setStatusDone() {
	logger.Debugf("transaction %v: setting status to 'done'", t.id)
	t.updated = GetCurrentTime()
	t.status = statusDone
}

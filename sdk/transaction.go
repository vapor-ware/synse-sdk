package sdk

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/xid"
	"github.com/vapor-ware/synse-server-grpc/go"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// InvalidateTransactionCache is a Test mechanism to invalidate the transaction
// cache. Returns an error if there is no transaction cache.
func InvalidateTransactionCache() (err error) {
	err = nil
	if transactionCache == nil {
		err = status.Errorf(codes.FailedPrecondition,
			"transactionCache not initialized.")
	} else {
		transactionCache = nil
	}
	return err
}

// SetupTransactionCache creates the transaction cache with the TTL in seconds.
// This needs to be called in order to have transactions.
// If this is called more than once, the cache is reinitialized and an error is
// returned. The caller may choose to ignore the error.
func SetupTransactionCache(ttl time.Duration) (err error) {
	err = nil
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

// NewTransaction creates a new Transaction instance. Upon creation, the
// Transaction is given a unique ID and is added to the transaction cache.
// This call will fail if the transaction cache is not initialized.
func NewTransaction() (*Transaction, error) {
	if transactionCache == nil {
		return nil, status.Errorf(codes.FailedPrecondition,
			"transactionCache not initialized. Call SetupTransactionCache.")
	}

	id := xid.New().String()
	now := GetCurrentTime()
	transaction := Transaction{
		id:      id,
		status:  statusUnknown,
		state:   stateOk,
		created: now,
		updated: now,
		message: "",
	}
	transactionCache.Set(id, &transaction, cache.DefaultExpiration)
	return &transaction, nil
}

// GetTransaction looks up the given transaction ID in the cache. If it exists,
// that Transaction is returned; otherwise nil is returned.
// This call will fail if the transaction cache is not initialized.
func GetTransaction(id string) (*Transaction, error) {
	if transactionCache == nil {
		return nil, status.Errorf(codes.FailedPrecondition,
			"transactionCache not initialized. Call SetupTransactionCache.")
	}

	transaction, found := transactionCache.Get(id)
	if found {
		return transaction.(*Transaction), nil
	}
	return nil, nil
}

// Transaction represents an asynchronous write transaction for the Plugin. It
// tracks the state and status of that transaction over its lifetime.
type Transaction struct {
	id      string
	status  synse.WriteResponse_WriteStatus
	state   synse.WriteResponse_WriteState
	created string
	updated string
	message string
}

// encode translates the Transaction to a corresponding gRPC WriteResponse.
func (t *Transaction) encode() *synse.WriteResponse {
	return &synse.WriteResponse{
		Status:  t.status,
		State:   t.state,
		Created: t.created,
		Updated: t.updated,
		Message: t.message,
	}
}

// setStateOk sets the Transaction to be in the 'ok' state. Since a pointer
// to this struct is stored in the cache, and update here should update the
// in-memory cache as well.
func (t *Transaction) setStateOk() {
	t.updated = GetCurrentTime()
	t.state = stateOk
}

// setStateError sets the Transaction to be in the 'error' state. Since a
// pointer to this struct is stored in the cache, and update here should
// update the in-memory cache as well.
func (t *Transaction) setStateError() {
	t.updated = GetCurrentTime()
	t.state = stateError
}

// setStatusUnknown sets the Transaction status to 'unknown'. Since a
// pointer to this struct is stored in the cache, and update here should
// update the in-memory cache as well.
func (t *Transaction) setStatusUnknown() {
	t.updated = GetCurrentTime()
	t.status = statusUnknown
}

// setStatusPending sets the Transaction status to 'pending'. Since a
// pointer to this struct is stored in the cache, and update here should
// update the in-memory cache as well.
func (t *Transaction) setStatusPending() {
	t.updated = GetCurrentTime()
	t.status = statusPending
}

// setStatusWriting sets the Transaction status to 'writing'. Since a
// pointer to this struct is stored in the cache, and update here should
// update the in-memory cache as well.
func (t *Transaction) setStatusWriting() {
	t.updated = GetCurrentTime()
	t.status = statusWriting
}

// setStatusDone sets the Transaction status to 'done'. Since a pointer
// to this struct is stored in the cache, and update here should update
// the in-memory cache as well.
func (t *Transaction) setStatusDone() {
	t.updated = GetCurrentTime()
	t.status = statusDone
}

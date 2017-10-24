package sdk

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/xid"
	synse "github.com/vapor-ware/synse-server-grpc/go"
)


const (

	// statuses

	// UNKNOWN is the unknown status, as defined by the gRPC API
	UNKNOWN = synse.WriteResponse_UNKNOWN

	// PENDING is the pending status, as defined by the gRPC API
	PENDING = synse.WriteResponse_PENDING

	// WRITING is the writing status, as defined by the gRPC API
	WRITING = synse.WriteResponse_WRITING

	// DONE is the done status, as defined by the gRPC API
	DONE = synse.WriteResponse_DONE

	// states

	// OK is the ok state, as defined by the gRPC API
	OK = synse.WriteResponse_OK

	// ERROR is the error state, as defined by the gRPC API
	ERROR = synse.WriteResponse_ERROR
)

// TransactionState is used to represent a Write transaction as it is
// processed. It holds the ID of the transaction, as well as the processing
// status and overall state. This model is used to look up transaction ids
// and return the asynchronous write responses.
type TransactionState struct {
	id        string
	status    synse.WriteResponse_WriteStatus
	state     synse.WriteResponse_WriteState
	created   string
	updated   string
	message   string
}

// toGRPC is a convenience method which translates the SDK's TransactionState
// model to the GRPC WriteResponse model.
func (t *TransactionState) toGRPC() *synse.WriteResponse {
	return &synse.WriteResponse{
		Status: t.status,
		State: t.state,
		Created: t.created,
		Updated: t.updated,
		Message: t.message,
	}
}


// transactionCache is a cache with a configurable default expiration time that
// is used to track the asynchronous write transactions as they are processed.
var transactionCache = cache.New(
	time.Duration(Config.Settings.Transaction.TTL)*time.Second,
	time.Duration(Config.Settings.Transaction.TTL)*2*time.Second,
)


// NewTransactionID creates a new `TransactionState` instance and gives it a
// new id. The newly created `TransactionState` is then added to the
// transactionCache.
func NewTransactionID() *TransactionState {
	id := xid.New().String()
	now := time.Now().String()

	ts := &TransactionState{id, UNKNOWN, OK, now, now, ""}
	transactionCache.Set(id, ts, cache.DefaultExpiration)
	return ts
}


// GetTransaction takes a transaction id and returns the corresponding
// `TransactionState`, if it exists. If no transaction is found, nil
// is returned.
func GetTransaction(id string) *TransactionState {
	transaction, found := transactionCache.Get(id)
	if found {
		return transaction.(*TransactionState)
	}
	return nil
}


// UpdateTransactionState updates the state field for a transaction's corresponding
// TransactionState. If the given id does not correspond to a known transaction,
// no action is taken.
func UpdateTransactionState(id string, state synse.WriteResponse_WriteState) {
	transaction, found := transactionCache.Get(id)
	if !found {
		Logger.Debugf("Did not find transaction %v when updating state.", id)
		return
	}

	t := transaction.(*TransactionState)
	now := time.Now().String()

	t.state = state
	t.updated = now
}


// UpdateTransactionStatus updates the status field for a transaction's corresponding
// TransactionState. If the given id does not correspond to a known transaction,
// no action is taken.
func UpdateTransactionStatus(id string, status synse.WriteResponse_WriteStatus) {
	transaction, found := transactionCache.Get(id)
	if !found {
		Logger.Debugf("Did not find transaction %v when updating status.", id)
		return
	}

	t := transaction.(*TransactionState)
	now := time.Now().String()

	t.status = status
	t.updated = now
}

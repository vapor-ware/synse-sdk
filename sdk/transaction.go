package sdk

import (
	"time"

	"github.com/rs/xid"
	//synse "github.com/vapor-ware/synse-server-grpc/go"
	synse "./synse"
)


const (
	// statuses
	UNKNOWN = synse.WriteResponse_UNKNOWN
	PENDING = synse.WriteResponse_PENDING
	WRITING = synse.WriteResponse_WRITING
	DONE = synse.WriteResponse_DONE

	// states
	OK = synse.WriteResponse_OK
	ERROR = synse.WriteResponse_ERROR
)

type TransactionState struct {
	id        string
	status    synse.WriteResponse_WriteStatus
	state     synse.WriteResponse_WriteState
	created   string
	updated   string
}


// FIXME - transaction expiration
// FIXME - on update, etc. add checks for if id not in transaction map

//
var transactionMap = make(map[string]*TransactionState)


//
func NewTransactionId() *TransactionState {
	id := xid.New().String()
	now := time.Now().String()

	ts := &TransactionState{id, UNKNOWN, OK, now, now}
	transactionMap[id] = ts
	return ts
}


//
func GetTransaction(id string) *TransactionState {
	return transactionMap[id]
}


//
func UpdateTransactionState(id string, state synse.WriteResponse_WriteState) {
	transaction := transactionMap[id]
	transaction.state = state
}


//
func UpdateTransactionStatus(id string, status synse.WriteResponse_WriteStatus) {
	transaction := transactionMap[id]
	transaction.status = status
}

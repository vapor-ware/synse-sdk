package sdk

import (
	"testing"
)


func TestNewTransactionId(t *testing.T) {
	ts := NewTransactionID()

	if ts.status != UNKNOWN {
		t.Error("New transaction should have status UNKNOWN.")
	}

	if ts.state != OK {
		t.Error("New transaction should be in the OK state.")
	}

	if ts.updated != ts.created {
		t.Error("The last update time and created time should be the same.")
	}

	if ts.message != "" {
		t.Error("A new transaction should not have a message.")
	}
}


func TestGetTransaction(t *testing.T) {
	// first create a new transaction
	t1 := NewTransactionID()

	// now, get the transaction.
	t2 := GetTransaction(t1.id)

	if t1 != t2 {
		t.Error("The two transactions should be the same.")
	}
}


func TestGetTransaction2(t *testing.T) {
	ts := GetTransaction("123")
	if ts != nil {
		t.Error("Transaction with id 123 should not exist.")
	}
}


func TestUpdateTransactionState(t *testing.T) {
	// first create a new transaction
	t1 := NewTransactionID()

	if t1.state != OK {
		t.Error("New transaction should be in the OK state.")
	}

	UpdateTransactionState(t1.id, ERROR)

	if t1.state != ERROR {
		t.Error("Failed to update transaction state to ERROR.")
	}
}


func TestUpdateTransactionState2(t *testing.T) {
	// nothing should happen for update of invalid id
	UpdateTransactionState("123", ERROR)
}


func TestUpdateTransactionStatus(t *testing.T) {
	// first create a new transaction
	t1 := NewTransactionID()

	if t1.status != UNKNOWN {
		t.Error("New transaction should have status UNKNOWN")
	}

	UpdateTransactionStatus(t1.id, DONE)

	if t1.status != DONE {
		t.Error("Failed to update transaction status to DONE")
	}
}


func TestUpdateTransactionStatus2(t *testing.T) {
	// nothing should happen for update of invalid id
	UpdateTransactionStatus("123", DONE)
}
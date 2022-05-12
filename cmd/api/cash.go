package main

import (
	"sync"
)

//Store = DB
type Store struct {
	mu   sync.RWMutex
	data map[string]string
}

//Create creates new Store by putting the empty map
func Create() *Store {
	s := &Store{
		data: make(map[string]string),
	}
	return s
}

//Transaction holds current transaction data
type Transaction struct {
	s        *Store // reference to database
	writable bool   // whether transaction is for Reading or Writing
}

//Set is to set the value to data
func (tx *Transaction) Set(key, value string) {
	tx.s.data[key] = value
}

//Get is to get the value from data
func (tx *Transaction) Get(key string) string {
	return tx.s.data[key]
}

//lock checks does it writable if true it locks to write else locks to read
func (tx *Transaction) lock() {
	if tx.writable {
		tx.s.mu.Lock()
	} else {
		tx.s.mu.RLock()
	}
}

//unlock checks does it writable if true it unlocks to write else unlocks to read
func (tx *Transaction) unlock() {
	if tx.writable {
		tx.s.mu.Unlock()
	} else {
		tx.s.mu.RUnlock()
	}
}

func (s *Store) Begin(writable bool) (*Transaction, error) {
	tx := &Transaction{
		s:        s,
		writable: writable,
	}
	tx.lock()

	return tx, nil
}

//managed Begins the Transactions
func (s *Store) managed(writable bool, fn func(tx *Transaction) error) (err error) {
	var tx *Transaction
	tx, err = s.Begin(writable)
	if err != nil {
		return
	}

	defer func() {
		if writable {
			tx.unlock()
		} else {
			tx.unlock()
		}
	}()

	err = fn(tx)
	return
}

//View gets data
func (s *Store) View(fn func(tx *Transaction) error) error {
	return s.managed(false, fn)
}

//Update sets data
func (s *Store) Update(fn func(tx *Transaction) error) error {
	return s.managed(true, fn)
}

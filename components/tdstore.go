package components

import (
	"sync"
	"errors"
	"database/sql"
	_ "github.com/lib/pq"
	"time"
	"fmt"
)

type TDStoreInterface interface {
	Set(item TDItem) (string, error)
	Get(key string) (TDItem, error)
	Truncate() error
	Dump() ([]TDItem, error)
	Close()
}

//Structs and functions relating to a RAMTDStore
type SafeTDStore struct {
	sync.RWMutex
	Items map[string]string
}

type RAMTDStore struct {
	I SafeTDStore
}

func (r *RAMTDStore) Set(item TDItem) (string, error) {
	r.I.Lock()
	defer r.I.Unlock()

	for k, v := range r.I.Items {
		if v == item.URL {
			return k, sql.ErrNoRows
		}
	}

	r.I.Items[item.Key] = item.URL
	return item.Key, nil

}

func (r *RAMTDStore) Get(key string) (TDItem, error) {
	r.I.RLock()
	defer r.I.RUnlock()

	if v, ok := r.I.Items[key]; ok {
		return TDItem{URL:v, Key: key}, nil
	} else {
		return TDItem{}, errors.New(fmt.Sprintf("key %s not found", key))
	}
}

func (r *RAMTDStore) Truncate() error {
	r.I.Lock()
	defer r.I.Unlock()
	r.I.Items = map[string]string{}

	return nil
}

func (r *RAMTDStore) Dump() ([]TDItem, error) {
	r.I.Lock()
	defer r.I.Unlock()
	var returns []TDItem

	for key, url := range r.I.Items {
		returns = append(returns, TDItem{Key: key, URL: url})
	}
	return returns, nil
}

func (r *RAMTDStore) Close() {
	r.Truncate()
}

func GenerateNewRAMTDS() TDStoreInterface {
	return &RAMTDStore{
		I: SafeTDStore{
			Items:map[string]string{},
		},
	}
}


//Structs and functions relating to a Postgres TDStore (PGTDStore)
type PGTDStore struct {
	db *sql.DB
}

func (p *PGTDStore) Set(item TDItem) (string, error) {
	if err := p.db.Ping(); err != nil {return "", err}
	var retKey string
	err := p.db.QueryRow(`INSERT INTO lard(k,u) VALUES ($1, $2) ON CONFLICT(u) DO UPDATE SET last_conflict=NOW() RETURNING k`, item.Key, item.URL).Scan(&retKey)
	if err != nil {return "", err}
	return retKey, nil
}
func (p *PGTDStore) Get(key string) (TDItem, error) {
	rI := &TDItem{
		Key: key,
	}

	if err := p.db.Ping(); err != nil {return *rI, err}
	err := p.db.QueryRow(`SELECT u FROM lard where k = $1`, key).Scan(&rI.URL)

	if err != nil {return *rI, err}

	return *rI, nil
}
func (p *PGTDStore) Truncate() error {
	return errors.New("plz don't")
}
func (p *PGTDStore) Dump() ([]TDItem, error) {
	return nil, errors.New("to be implemented maybe, probably not")
}
func (p *PGTDStore) Close() {
	p.db.Close()
}

func GenerateNewPGTDS(dsn string) TDStoreInterface {
	var err error
	d, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	if err = d.Ping(); err != nil {
		panic(err)
	}
	d.SetMaxIdleConns(1)
	d.SetMaxOpenConns(100)
	d.SetConnMaxLifetime(5 * time.Second)
	return &PGTDStore{
		db:d,
	}
}
package goodies

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Persister type performing reccurent persists
type Persister struct {
	*Storage
	stop     chan bool
	filename string
	interval time.Duration
}

type StoppableProvider interface {
	Provider
	Stop()
}

//NewGoodiesPersistedStorage Creates an instance of persisted goodies storage
func NewGoodiesPersistedStorage(ttl time.Duration, filename string, persistenceInterval time.Duration) StoppableProvider {

	if filename == "" {
		panic("Filename cannot be empty")
	}
	storage := NewGoodiesStorage(ttl)

	path := filepath.Dir(filename)
	os.MkdirAll(path, os.ModePerm)

	persisted := Persister{
		Storage:  storage,
		stop:     make(chan bool),
		filename: filename,
		interval: persistenceInterval,
	}
	initialStorage := make(map[string]goodiesItem)
	_ = persisted.Load(&initialStorage)
	persisted.storage = initialStorage
	go persisted.runPersister()
	return persisted
}

//Stop method is a nice way to clearly stop the cache
func (p Persister) Stop() {
	p.stop <- true
}

func (p *Persister) runPersister() {
	persistTrigger := time.NewTicker(p.interval)
	for {
		select {
		case <-persistTrigger.C:
			p.cleanupOutdated()
			if err := p.Save(p.getBlob()); err != nil {
				fmt.Printf("Backup not saved: %v\n", err)
			}
		case <-p.stop:
			p.cleanupOutdated()
			if err := p.Save(p.getBlob()); err != nil {
				fmt.Printf("Backup not saved: %v\n", err)
			}
			return
		}
	}
}

// Load Load blob from file storage
func (p *Persister) Load(data interface{}) error {
	file, err := os.Open(p.filename)
	if err == nil {
		decoder := gob.NewDecoder(file)
		err = decoder.Decode(data)
	}
	file.Close()
	return err
}

// Save Save blob to file storage
func (p *Persister) Save(data interface{}) error {
	file, err := os.Create(p.filename)
	if err == nil {
		encoder := gob.NewEncoder(file)
		err = encoder.Encode(data)
	}
	file.Close()
	return err
}

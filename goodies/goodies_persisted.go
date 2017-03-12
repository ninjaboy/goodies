package goodies

import "encoding/gob"

import "time"
import "os"

// Persister type performing reccurent persists
type Persister struct {
	stop     chan<- bool
	filename string
	interval time.Duration
}

// NewPersister Creates file based persister
func NewPersister(filename string, interval time.Duration) *Persister {
	persister := &Persister{
		stop:     make(chan bool),
		filename: filename,
		interval: interval,
	}
	return persister
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

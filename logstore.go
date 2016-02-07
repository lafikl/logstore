package logstore

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"os"
	"sync"
)

// ErrNoPartition is returned when the requested partition doesn't exist
var ErrNoPartition = errors.New("Partition doesn't exist")

// ErrShortPayload returned when UnMarshal is given a too small buffer of messages
var ErrShortPayload = errors.New("Too short payload of messages")

// LogStore is an immutable append-only storage engine for logs.
type LogStore struct {
	dataDir    string
	numParts   int
	Partitions []*Partition
}

// Partition bla bla
type Partition struct {
	*sync.RWMutex
	Fd  *os.File
	Idx int
}

// Message is a representation of what gets stored on disk
type Message struct {
	Length   uint64
	Checksum uint32
	Payload  []byte
}

// New returns a new LogStore instance
func New(dataDir string, partitions int) *LogStore {
	l := &LogStore{dataDir, partitions, []*Partition{}}
	l.Setup()
	return l
}

func (l *LogStore) newPartition(idx int, fd *os.File) *Partition {
	return &Partition{&sync.RWMutex{}, fd, idx}
}

// Setup makes sure that all the files are in place
func (l *LogStore) Setup() {
	os.Mkdir(l.dataDir, 0764)
	for i := 0; i < l.numParts; i++ {
		fname := fmt.Sprintf("%s%d.bilog", l.dataDir, i)
		if fd, error := os.OpenFile(fname, os.O_APPEND|os.O_RDWR, os.ModeAppend); error == nil {
			l.Partitions = append(l.Partitions, l.newPartition(i, fd))
			continue
		}
		_, err := os.Create(fname)
		if err != nil {
			panic(err)
		}
		fd, err := os.OpenFile(fname, os.O_APPEND|os.O_RDWR, os.ModeAppend)
		if err != nil {
			panic(err)
		}
		l.Partitions = append(l.Partitions, l.newPartition(i, fd))
	}
}

// Append appends the given message to the log file
func (l *LogStore) Append(pkey string, payload []byte) (n int64, err error) {
	fn := fnv.New32a()
	fn.Write([]byte(pkey))
	hash := fn.Sum32()
	idx := hash % uint32(l.numParts-1)
	prtn := l.Partitions[idx]
	msg := l.msgify(payload)
	prtn.Lock()
	defer prtn.Unlock()
	n, err = msg.WriteTo(prtn.Fd)
	return n, err
}

// Partition returns a refernce for the requested partition
func (l *LogStore) Partition(idx int) (*Partition, error) {
	if idx > l.numParts {
		return nil, ErrNoPartition
	}
	prtn := l.Partitions[idx]
	return prtn, nil
}

// UnMarshal deserializes the given byte slice into Messages
func (l *LogStore) UnMarshal(b []byte) ([]*Message, error) {
	msgs := []*Message{}
	blen := len(b)
	if blen < 13 {
		return nil, ErrShortPayload
	}
	s := 0
	for {
		msg := &Message{}
		n, err := l.parse(b[s:], msg)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
		s += n
		if s == blen {
			break
		}

	}
	return msgs, nil
}

func (l *LogStore) parse(payload []byte, msg *Message) (n int, err error) {
	ln := binary.BigEndian.Uint64(payload[:8])
	crc := binary.BigEndian.Uint32(payload[8:12])
	// TODO(KL): verify message against checksum
	m := payload[12:ln]
	msg.Length = ln
	msg.Payload = m
	msg.Checksum = crc
	return int(ln), nil
}

func (l *LogStore) Read(partition int, offset int64, buf []byte) (n int, err error) {
	if partition > l.numParts {
		return 0, ErrNoPartition
	}
	prtn := l.Partitions[partition]
	prtn.RLock()
	defer prtn.RUnlock()
	prtn.Fd.Seek(offset, 0)
	rdr := bufio.NewReader(prtn.Fd)
	n, err = rdr.Read(buf)
	return
}

func (l *LogStore) msgify(payload []byte) *bytes.Buffer {
	buff := new(bytes.Buffer)
	// length/checksum+payload (8) + checksum (4) + N payload
	binary.Write(buff, binary.BigEndian, uint64(8+4+len(payload)))
	binary.Write(buff, binary.BigEndian, checksum(payload))
	binary.Write(buff, binary.BigEndian, payload)
	return buff
}

func checksum(data []byte) uint32 {
	tbl := crc32.MakeTable(0x04C11DB7)
	return crc32.Checksum(data, tbl)
}

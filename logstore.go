package logstore

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"os"
)

// LogStore is an immutable append-only storage engine for logs.
type LogStore struct {
	dataDir  string
	numParts int
	fdParts  []*os.File
}

// New returns a new LogStore instance
func New(dataDir string, partitions int) *LogStore {
	l := &LogStore{dataDir, partitions, []*os.File{}}
	l.Setup()
	return l
}

// Setup makes sure that all the files are in place
func (l *LogStore) Setup() {
	os.Mkdir(l.dataDir, 0777)
	for i := 0; i < l.numParts; i++ {
		fname := fmt.Sprintf("%s%d.bilog", l.dataDir, i)
		if fd, error := os.OpenFile(fname, os.O_APPEND|os.O_WRONLY, os.ModeAppend); error == nil {
			l.fdParts = append(l.fdParts, fd)
			continue
		}
		_, err := os.Create(fname)
		if err != nil {
			panic(err)
		}
		fd, err := os.OpenFile(fname, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			panic(err)
		}
		l.fdParts = append(l.fdParts, fd)
	}
}

// Append appends the given message to the log file
func (l *LogStore) Append(pkey string, payload []byte) error {
	fn := fnv.New32a()
	fn.Write([]byte(pkey))
	hash := fn.Sum32()
	prtn := hash % uint32(l.numParts-1)
	fd := l.fdParts[prtn]
	msg := l.msgify(payload)
	fmt.Println(prtn)
	msg.WriteTo(fd)
	return nil
}

func (l *LogStore) msgify(payload []byte) *bytes.Buffer {
	buff := new(bytes.Buffer)
	// length/checksum+payload (8) + checksum (4) + N payload
	binary.Write(buff, binary.LittleEndian, uint64(4+len(payload)))
	binary.Write(buff, binary.LittleEndian, checksum(payload))
	binary.Write(buff, binary.LittleEndian, payload)
	return buff
}

func checksum(data []byte) uint32 {
	tbl := crc32.MakeTable(0x04C11DB7)
	return crc32.Checksum(data, tbl)
}

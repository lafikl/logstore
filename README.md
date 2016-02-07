# LogStore
A partitioned & immutable append-only storage engine, meant for logs.


# Why
It's meant for use with any system that wants to implement an append-only data structure that can be consumed in parallel by multiple consumers.
The end goal is to use as backing store for a message queue i'm working on currently.
Which will provide the distributed layer on top of it.

# Features
- Very fast writes, because it only appends to the end of the partition.
- Very fast read, because it only reads sequentially.
- Support multiple concurrent readers via partitioning the log files.
- Guarantees order on the partition level only. Which eliminates the need for consensus.

# Use Cases
- Commit Log
- Event Sourcing
- Binary Web Logs
- Message Queues

# Example

```go
    import (
        "fmt"
        "log"

        "github.com/lafikl/logstore"
    )

    func main() {
        lstore := logstore.New("logs/", 10)
        # appending logs
        pkey := "10084"
        lstore.Append(pkey, []byte("Hello World!"))

        // fetching a partition
        // index starts from 0.
        prtn, err := lstore.Partition(5)
        if err != nil {
            log.Fatal(err)
        }
        // hint:
        // you can use the prtn.Fd to send messages slices to other LogStore machines without copying data into use-space
        fmt.Println(prtn.Idx)

        // reading logs
        buff := make([]byte, 2024)
        partition := 0
        offset := int64(0)
        n, err := lstore.Read(partition, offset, buff)
        if err != nil {
            log.Fatal(err)
        }
        msgs, err := l.UnMarshal(b[:n])
        if err != nil {
            log.Fatal(err)
        }
        for idx, msgs := range msgs {
            fmt.Println(idx, string(msgs.Payload))
        }
    }
```

# API Docs
[GoDoc](https://godoc.org/github.com/lafikl/logstore)

# Status
Still in a very early state. Don't use it with anything precious.


# License
The MIT License (MIT)

Copyright (c) 2016 Khalid Lafi

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

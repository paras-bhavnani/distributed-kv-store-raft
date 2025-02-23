# Lab 2: Distributed, Fault-Tolerant Key-Value Store (MIT 6.5840)

This branch contains the implementation of **Lab 2** of MIT's 6.5840 Distributed Systems course. The goal of this lab is to build a single-server key/value store that ensures **linearizability** and handles **exactly-once semantics** for client operations, even in the presence of network failures.

---

## **Overview**

The key-value server supports three RPC operations:
- **Put(key, value)**: Installs or replaces the value for a given key.
- **Append(key, arg)**: Appends a string argument (`arg`) to an existing value for a key.
- **Get(key)**: Retrieves the current value associated with a key.

### **Key Features**
1. **Linearizability**: Ensures that operations appear as if they were executed one at a time in some sequential order.
2. **At-least-once Semantics with Duplicate Detection**: Handles duplicate RPCs and retries by clients without executing operations multiple times.
3. **Fault Tolerance**: Designed to handle unreliable networks (e.g., dropped messages).

---

## **Project Structure**

The implementation is contained in the following files:
- `kvsrv/client.go`: Contains the `Clerk` implementation for interacting with the server via RPC.
- `kvsrv/server.go`: Implements the key-value store server and its RPC handlers.
- `kvsrv/common.go`: Contains shared data structures and constants.
- `kvsrv/config.go`: Configuration setup for testing provided by MIT.
- `kvsrv/test_test.go`: Test cases provided by MIT to validate correctness.

---

## **Getting Started**

### Prerequisites
- Go programming language installed (version ≥ 1.20).
- A Unix-like environment (Linux, macOS, or WSL2 on Windows).

### Setup Instructions
1. Clone this repository:

``` bash
git clone https://github.com/paras-bhavnani/distributed-kv-store-raft.git
cd distributed-kv-store-raft
git checkout lab2
```

2. Navigate to the `src/kvsrv` directory:

``` bash
cd src/kvsrv
```

3. Run tests to validate your implementation:

``` bash
go test
```

---

## Implementation Tasks

1. Implement RPC-sending code in `Clerk` methods (`Get`, `Put`, `Append`) in `client.go`.
2. Implement RPC handlers (`Get`, `PutAppend`) in `server.go`.
3. Add unique request identification (ClientID and RequestID/SequenceNum) to RPCs.
4. Implement duplicate detection logic in `server.go` to ensure at-least-once semantics.
5. Update `Clerk` logic in `client.go` to retry requests if replies are not received.

---

## **Testing**

Run the provided test suite to validate your implementation:

``` bash
go test
```

My Final output after completing all tasks successfully:

``` bash
Test: one client ...
  ... Passed -- t  3.7 nrpc 29451 ops 29451
Test: many clients ...
  ... Passed -- t  4.8 nrpc 87068 ops 87068
Test: unreliable net, many clients ...
Duplicated

Duplicated

Duplicated
x 2 0 yx 2 1 yx 2 2 yx 2 3 yx 2 4 yx 2 5 yx 2 6 y
Duplicated
x 3 0 yx 3 1 yx 3 2 yx 3 3 y
Duplicated
x 1 0 yx 1 1 yx 1 2 yx 1 3 yx 1 4 yx 1 5 yx 1 6 yx 1 7 yx 1 8 yx 1 9 yx 1 10 yx 1 11 yx 1 12 yx 1 13 yx 1 14 y
Duplicated
x 1 0 yx 1 1 yx 1 2 yx 1 3 yx 1 4 yx 1 5 yx 1 6 yx 1 7 yx 1 8 yx 1 9 yx 1 10 yx 1 11 yx 1 12 yx 1 13 yx 1 14 y
Duplicated
x 4 0 yx 4 1 y
Duplicated
x 1 0 yx 1 1 yx 1 2 yx 1 3 y
Duplicated
x 3 0 yx 3 1 yx 3 2 yx 3 3 y
Duplicated
x 1 0 yx 1 1 yx 1 2 yx 1 3 yx 1 4 yx 1 5 yx 1 6 y
Duplicated
x 1 0 yx 1 1 yx 1 2 yx 1 3 yx 1 4 yx 1 5 yx 1 6 yx 1 7 y
Duplicated
x 3 0 yx 3 1 yx 3 2 yx 3 3 yx 3 4 yx 3 5 yx 3 6 yx 3 7 yx 3 8 yx 3 9 yx 3 10 yx 3 11 yx 3 12 y
Duplicated
x 3 0 yx 3 1 yx 3 2 yx 3 3 yx 3 4 yx 3 5 yx 3 6 yx 3 7 yx 3 8 yx 3 9 yx 3 10 yx 3 11 yx 3 12 y
Duplicated
x 0 0 y
Duplicated
x 2 0 y
Duplicated
x 1 0 yx 1 1 yx 1 2 yx 1 3 yx 1 4 y
Duplicated
x 2 0 yx 2 1 y
Duplicated
x 4 0 yx 4 1 yx 4 2 yx 4 3 yx 4 4 yx 4 5 yx 4 6 yx 4 7 yx 4 8 yx 4 9 yx 4 10 yx 4 11 y
Duplicated
x 3 0 yx 3 1 yx 3 2 yx 3 3 yx 3 4 yx 3 5 yx 3 6 yx 3 7 yx 3 8 yx 3 9 yx 3 10 y
Duplicated
x 1 0 yx 1 1 yx 1 2 yx 1 3 yx 1 4 yx 1 5 yx 1 6 yx 1 7 yx 1 8 y
  ... Passed -- t  4.1 nrpc   497 ops  395
Test: concurrent append to same key, unreliable ...
Duplicated
x 4 0 yx 4 1 yx 1 0 yx 0 0 yx 2 0 yx 0 1 yx 4 2 yx 0 2 yx 0 3 y
Duplicated
x 4 0 yx 4 1 yx 1 0 yx 0 0 yx 2 0 yx 0 1 yx 4 2 yx 0 2 yx 0 3 y
Duplicated
x 4 0 yx 4 1 yx 1 0 yx 0 0 yx 2 0 yx 0 1 yx 4 2 yx 0 2 yx 0 3 yx 1 1 yx 2 1 yx 4 3 yx 4 4 yx 0 4 yx 0 5 yx 0 6 yx 0 7 y
Duplicated
x 4 0 yx 4 1 yx 1 0 yx 0 0 yx 2 0 yx 0 1 yx 4 2 yx 0 2 yx 0 3 y
Duplicated
x 4 0 yx 4 1 yx 1 0 yx 0 0 yx 2 0 yx 0 1 yx 4 2 yx 0 2 yx 0 3 yx 1 1 yx 2 1 yx 4 3 yx 4 4 yx 0 4 yx 0 5 yx 0 6 yx 0 7 yx 2 2 yx 4 5 yx 0 8 yx 4 6 yx 4 7 yx 0 9 yx 4 8 yx 3 0 yx 4 9 yx 3 1 yx 3 2 yx 3 3 yx 3 4 yx 2 3 yx 2 4 yx 2 5 yx 2 6 yx 2 7 yx 1 2 yx 2 8 y
Duplicated
x 4 0 yx 4 1 yx 1 0 yx 0 0 yx 2 0 yx 0 1 yx 4 2 yx 0 2 yx 0 3 yx 1 1 yx 2 1 yx 4 3 yx 4 4 yx 0 4 yx 0 5 yx 0 6 yx 0 7 yx 2 2 yx 4 5 yx 0 8 yx 4 6 yx 4 7 yx 0 9 yx 4 8 yx 3 0 yx 4 9 yx 3 1 yx 3 2 yx 3 3 yx 3 4 yx 2 3 yx 2 4 yx 2 5 yx 2 6 yx 2 7 yx 1 2 yx 2 8 yx 1 3 yx 2 9 yx 3 5 yx 3 6 yx 3 7 y
Duplicated
x 4 0 yx 4 1 yx 1 0 yx 0 0 yx 2 0 yx 0 1 yx 4 2 yx 0 2 yx 0 3 yx 1 1 yx 2 1 yx 4 3 yx 4 4 yx 0 4 yx 0 5 yx 0 6 yx 0 7 yx 2 2 yx 4 5 yx 0 8 yx 4 6 yx 4 7 yx 0 9 yx 4 8 yx 3 0 yx 4 9 yx 3 1 yx 3 2 yx 3 3 yx 3 4 yx 2 3 yx 2 4 yx 2 5 yx 2 6 yx 2 7 yx 1 2 yx 2 8 yx 1 3 yx 2 9 yx 3 5 yx 3 6 yx 3 7 y
Duplicated
x 4 0 yx 4 1 yx 1 0 yx 0 0 yx 2 0 yx 0 1 yx 4 2 yx 0 2 yx 0 3 yx 1 1 yx 2 1 yx 4 3 yx 4 4 yx 0 4 yx 0 5 yx 0 6 yx 0 7 yx 2 2 yx 4 5 yx 0 8 yx 4 6 yx 4 7 yx 0 9 yx 4 8 yx 3 0 yx 4 9 yx 3 1 yx 3 2 yx 3 3 yx 3 4 yx 2 3 yx 2 4 yx 2 5 yx 2 6 yx 2 7 yx 1 2 yx 2 8 yx 1 3 yx 2 9 yx 3 5 yx 3 6 yx 3 7 yx 3 8 y
Duplicated
x 4 0 yx 4 1 yx 1 0 yx 0 0 yx 2 0 yx 0 1 yx 4 2 yx 0 2 yx 0 3 yx 1 1 yx 2 1 yx 4 3 yx 4 4 yx 0 4 yx 0 5 yx 0 6 yx 0 7 yx 2 2 yx 4 5 yx 0 8 yx 4 6 yx 4 7 yx 0 9 yx 4 8 yx 3 0 yx 4 9 yx 3 1 yx 3 2 yx 3 3 yx 3 4 yx 2 3 yx 2 4 yx 2 5 yx 2 6 yx 2 7 yx 1 2 yx 2 8 yx 1 3 yx 2 9 yx 3 5 yx 3 6 yx 3 7 yx 3 8 yx 1 4 yx 3 9 yx 1 5 yx 1 6 y
  ... Passed -- t  1.0 nrpc    66 ops   52
Test: memory use get ...
  ... Passed -- t  1.7 nrpc     5 ops    0
Test: memory use put ...
  ... Passed -- t  0.2 nrpc     2 ops    0
Test: memory use append ...
  ... Passed -- t  0.3 nrpc     2 ops    0
Test: memory use many puts ...
  ... Passed -- t  9.6 nrpc 100000 ops    0
Test: memory use many gets ...
  ... Passed -- t  9.1 nrpc 100001 ops    0
PASS
ok      github.com/paras-bhavnani/distributed-kv-store-raft/kvsrv       35.640s
```

## Current Status

All test cases have been successfully passed, including tests for concurrent operations and unreliable networks.

---

## **References**

This lab is part of MIT's [6.5840 Distributed Systems](http://nil.csail.mit.edu/6.5840/2024/labs/lab-kvsrv.html) course.

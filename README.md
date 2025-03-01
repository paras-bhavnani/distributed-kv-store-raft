# Lab 3: Raft Consensus Algorithm (MIT 6.5840)

This branch contains the implementation of Lab 3 of MIT's 6.5840 Distributed Systems course. The goal of this lab is to implement the Raft consensus algorithm, which will be used as the foundation for building a fault-tolerant key/value storage system in subsequent labs.

## Overview

Raft is a consensus algorithm designed to be easy to understand and implement. It provides a way for a cluster of servers to agree on a series of operations, even in the face of network partitions and server failures.

## Key Features

- **Leader Election**: Elects a leader among the servers to coordinate operations.
- **Log Replication**: Ensures all servers maintain the same log of operations.
- **Safety**: Guarantees that if any server has applied a particular log entry to its state machine, no other server will ever apply a different log entry for the same index.

## Project Structure

The implementation is contained in the following files:

- `raft/raft.go`: Contains the main Raft implementation.
- `raft/config.go`: Configuration setup for testing provided by MIT.
- `raft/persister.go`: Handles persistence of Raft state.
- `raft/test_test.go`: Test cases provided by MIT to validate correctness.

## Getting Started

### Prerequisites

- Go programming language installed (version ≥ 1.20).
- A Unix-like environment (Linux, macOS, or WSL2 on Windows).

### Setup Instructions

Clone this repository:

```bash
git clone https://github.com/paras-bhavnani/distributed-kv-store-raft.git
cd distributed-kv-store-raft
git checkout lab3
```

Navigate to the `src/raft` directory:

```bash
cd src/raft
```

Run tests to validate your implementation:

```bash
go test
```

## Implementation Tasks

### Part 3A: Leader Election

- Implemented Raft leader election and heartbeats (AppendEntries RPCs with no log entries).
- Ensured a single leader is elected and remains the leader if there are no failures.
- Implemented proper term handling and vote request logic.
- Handled leader failures and network partitions correctly.

### Part 3B: Log Replication

- Implemented log replication functionality.
- Ensured proper handling of AppendEntries RPCs for both heartbeats and log entries.
- Implemented log consistency check and conflict resolution.
- Added commit index management and log application to state machine.

## Testing

Run the provided test suite to validate your implementation:

```bash
go test -run 3A
go test -run 3B
```

## Current Status

Parts 3A (Leader Election) and 3B (Log Replication) have been implemented and all related tests have passed successfully.

## References

This lab is part of MIT's 6.5840 Distributed Systems course.


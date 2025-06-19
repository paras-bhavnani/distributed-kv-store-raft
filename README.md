# Lab 3: Raft Consensus Algorithm (MIT 6.5840)

This branch contains the implementation of Lab 3 of MIT's 6.5840 Distributed Systems course. The goal of this lab is to implement the Raft consensus algorithm, which will be used as the foundation for building a fault-tolerant key/value storage system in subsequent labs.

## Overview

Raft is a consensus algorithm designed to be easy to understand and implement. It provides a way for a cluster of servers to agree on a series of operations, even in the face of network partitions and server failures. It includes:

- **3A**: Leader Election
- **3B**: Log Replication
- **3C**: Persistence
- **3D**: Log Compaction with Snapshots (Complete)

## Key Features

- **Leader Election**: Elects a leader among the servers to coordinate operations.
- **Log Replication**: Ensures all servers maintain the same log of operations.
- **Persistence**: Survive crashes/restarts.
- **Snapshotting**: Efficient log compaction and recovery (Part 3D).

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

### Part 3C: Persistence (Hard)

#### Description

Task 3C required implementing persistence mechanisms so that Raft-based servers can recover from crashes and resume operation without losing consistency. This involved:

- **Persisting State**:
  - Persisted `currentTerm`, `votedFor`, and `log` as specified in Figure 2 of the Raft paper.
  - Used the `Persister` object provided by MIT's framework to save and restore persistent state.

- **Serialization**:
  - Used the `labgob` encoder/decoder for serializing and deserializing persistent state into byte arrays.
  - Implemented the `persist()` function to save state whenever it changed.
  - Implemented the `readPersist()` function to restore state during server initialization.

- **Log Recovery Optimization**:
  - Implemented optimizations for backing up `nextIndex` by more than one entry at a time during log recovery.
  - Handled rejection messages with additional metadata (`XTerm`, `XIndex`, `XLen`) to efficiently align follower logs with leaders.

### Part 3D: Log Compaction (Snapshots)

In 3D, Raft supports log compaction using **snapshots** to avoid unbounded log growth. When the service layer indicates, Raft discards old log entries and persists a compact snapshot of the state up to a given index. If a follower falls far behind (or a server restarts), the leader sends snapshots using the `InstallSnapshot` RPC rather than replaying an unbounded log. See [Section 7 of the extended Raft paper](https://raft.github.io/raft.pdf) for the protocol details.

Implemented:
- `Snapshot(index, snapshot []byte)` to trim logs and persist snapshots.
- `InstallSnapshot` RPC, supporting full state transfer and correct application.
- Careful log indexing and persistence, even after crashes/restarts.
- Memory management to enable Go’s GC to reclaim space.

**Lab 3 is now fully complete and passes all tests, including the hardest cases in 3D.**

## Testing

Run the provided test suite to validate your implementation:

Run all 3A–3D tests:

```bash
go test
```

or selectively:

```bash
go test -run 3A
go test -run 3B
go test -run 3C
go test -run 3D
```

### Testing after Implementation of 3D

Run the provided test suite multiple times to validate your implementation:

```bash
for i in {0..10}; do go test; done
```

Example output:

```plaintext
Test (3A): initial election ...
  ... Passed --   3.6  3   60   16440    0
Test (3A): election after network failure ...
  ... Passed --   5.5  3  122   23886    0
Test (3A): multiple elections ...
  ... Passed --   6.1  7  426   91010    0
Test (3B): basic agreement ...
  ... Passed --   0.9  3   16    4394    3
Test (3B): RPC byte count ...
  ... Passed --   1.8  3   48  113846   11
Test (3B): test progressive failure of followers ...
  ... Passed --   4.9  3  110   23617    3
Test (3B): test failure of leaders ...
  ... Passed --   5.3  3  180   38568    3
Test (3B): agreement after follower reconnects ...
  ... Passed --   5.7  3  122   32996    8
Test (3B): no agreement if too many followers disconnect ...
  ... Passed --   3.8  5  170   36682    3
Test (3B): concurrent Start()s ...
  ... Passed --   1.1  3   26    7491    6
Test (3B): rejoin of partitioned leader ...
  ... Passed --   4.3  3  136   31440    4
Test (3B): leader backs up quickly over incorrect follower logs ...
  ... Passed --  17.8  5 2013 1879356  102
Test (3B): RPC counts aren't too high ...
  ... Passed --   2.7  3   56   17010   12
Test (3C): basic persistence ...
  ... Passed --   4.8  3   86   21996    6
Test (3C): more persistence ...
  ... Passed --  16.5  5  832  186998   16
Test (3C): partitioned leader and one follower crash, leader restarts ...
  ... Passed --   2.1  3   32    8045    4
Test (3C): Figure 8 ...
  ... Passed --  34.0  5  776  321102   44
Test (3C): unreliable agreement ...
  ... Passed --   1.8  5 1035  347675  246
Test (3C): Figure 8 (unreliable) ...
  ... Passed --  33.8  5 10732 21685840  129
Test (3C): churn ...
  ... Passed --  16.1  5 8278 37921453 1859
Test (3C): unreliable churn ...
  ... Passed --  16.4  5 4444 12783448  954
Test (3D): snapshots basic ...
  ... Passed --   4.1  3  295  115787  220
Test (3D): install snapshots (disconnect) ...
  ... Passed --  43.2  3 1479  779614  362
Test (3D): install snapshots (disconnect+unreliable) ...
  ... Passed --  46.4  3 1609  688265  316
Test (3D): install snapshots (crash) ...
  ... Passed --  31.0  3 1044  545885  323
Test (3D): install snapshots (unreliable+crash) ...
  ... Passed --  35.0  3 1226  712549  332
Test (3D): crash and restart all servers ...
  ... Passed --   9.0  3  230   67420   52
Test (3D): snapshot initialization after crash ...
  ... Passed --   3.3  3   68   19572   14
PASS
ok      github.com/paras-bhavnani/distributed-kv-store-raft/raft        361.306s

  ... (runs 10 more times)
```

(See `lab3.log` for a full run.)

## Current Status

Parts 3A (Leader Election), 3B (Log Replication), 3C (Persistence), and 3D (Log Compaction) have been implemented and all related tests have passed successfully.

Test 3D now always passes:

```
go test -run 3D
Test (3D): snapshots basic ...
  ... Passed --   4.3  3  310  117957  213
Test (3D): install snapshots (disconnect) ...
  ... Passed --  42.9  3 1413  501127  295
Test (3D): install snapshots (disconnect+unreliable) ...
  ... Passed --  46.9  3 1588  584697  315
Test (3D): install snapshots (crash) ...
  ... Passed --  31.9  3 1078  428997  319
Test (3D): install snapshots (unreliable+crash) ...
  ... Passed --  38.5  3 1210  468749  298
Test (3D): crash and restart all servers ...
  ... Passed --   8.6  3  212   62022   47
Test (3D): snapshot initialization after crash ...
  ... Passed --   3.2  3   68   19468   14
PASS
ok      github.com/paras-bhavnani/distributed-kv-store-raft/raft        176.567s
```

## References

- This project is for MIT’s [6.5840 Distributed Systems](http://nil.csail.mit.edu/6.5840/2024/index.html) (Spring 2024), Lab 3
- [The Raft Paper](https://raft.github.io/raft.pdf)


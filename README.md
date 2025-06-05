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

## Testing

Run the provided test suite to validate your implementation:

```bash
go test -run 3A
go test -run 3B
go test -run 3C
```

### Testing after Implementation of 3C

Run the provided test suite multiple times to validate your implementation:

```bash
for i in {0..10}; do go test; done
```

Example output:

```plaintext
Test (3A): initial election ...
  ... Passed --   3.6  3   60   16092    0
Test (3A): election after network failure ...
  ... Passed --   5.1  3  118   22892    0
Test (3A): multiple elections ...
  ... Passed --   7.1  7  594  117844    0
Test (3B): basic agreement ...
  ... Passed --   0.8  3   14    3762    3
Test (3B): RPC byte count ...
  ... Passed --   1.8  3   48  113666   11
Test (3B): test progressive failure of followers ...
  ... Passed --   4.8  3  108   23239    3
Test (3B): test failure of leaders ...
  ... Passed --   5.4  3  181   38304    3
Test (3B): agreement after follower reconnects ...
  ... Passed --   4.1  3   92   23905    7
Test (3B): no agreement if too many followers disconnect ...
  ... Passed --   3.8  5  187   40212    3
Test (3B): concurrent Start()s ...
  ... Passed --   1.1  3   22    6069    6
Test (3B): rejoin of partitioned leader ...
  ... Passed --   7.3  3  196   45906    4
Test (3B): leader backs up quickly over incorrect follower logs ...
  ... Passed --  15.4  5 2735 2941478  102
Test (3B): RPC counts aren't too high ...
  ... Passed --   2.1  3   47   14182   12
Test (3C): basic persistence ...
  ... Passed --   4.3  3   78   19414    6
Test (3C): more persistence ...
  ... Passed --  19.4  5 1024  220592   17
Test (3C): partitioned leader and one follower crash, leader restarts ...
  ... Passed --   2.2  3   34    8397    4
Test (3C): Figure 8 ...
  ... Passed --  32.9  5  827  181905   38
Test (3C): unreliable agreement ...
  ... Passed --   1.8  5 1117  358753  246
Test (3C): Figure 8 (unreliable) ...
  ... Passed --  36.5  5 18357 58907792  143
Test (3C): churn ...
  ... Passed --  16.1  5 15450 125333403 2737
Test (3C): unreliable churn ...
  ... Passed --  16.1  5 6519 4637723 1168

  ... (runs 9 more times)
```

## Current Status

Parts 3A (Leader Election), 3B (Log Replication), and 3C (Persistence) have been implemented and all related tests have passed successfully.

Test 3D is failing intermittently, needs fixing current test run shows this:

```
go test -run 3D                                       
Test (3D): snapshots basic ...
  ... Passed --   5.7  3  192   73992  226
Test (3D): install snapshots (disconnect) ...
--- FAIL: TestSnapshotInstall3D (53.88s)
    config.go:605: one(3388814983182284110) failed to reach agreement
Test (3D): install snapshots (disconnect+unreliable) ...
  ... Passed --  51.8  3 1236  533454  331
Test (3D): install snapshots (crash) ...
  ... Passed --  38.1  3  734  347582  286
Test (3D): install snapshots (unreliable+crash) ...
  ... Passed --  42.9  3  834  443290  356
Test (3D): crash and restart all servers ...
--- FAIL: TestSnapshotAllCrash3D (12.76s)
    config.go:605: one(3681762713053052204) failed to reach agreement
Test (3D): snapshot initialization after crash ...
--- FAIL: TestSnapshotInit3D (13.53s)
    config.go:605: one(92410175302122921) failed to reach agreement
FAIL
exit status 1
FAIL    github.com/paras-bhavnani/distributed-kv-store-raft/raft        218.957s
```

## References

This lab is part of MIT's 6.5840 Distributed Systems course.


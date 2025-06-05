package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"bytes"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/paras-bhavnani/distributed-kv-store-raft/labgob"
	"github.com/paras-bhavnani/distributed-kv-store-raft/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 3D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 3D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

type LogEntry struct {
	Term    int
	Command interface{}
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (3A, 3B, 3C).

	currentTerm int
	votedFor    int
	log         []LogEntry

	state          string // "follower", "candidate", or "leader"
	electionTimer  *time.Timer
	heartbeatTimer *time.Timer
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	// Log replication state
	commitIndex int   // index of highest log entry known to be committed
	lastApplied int   // index of highest log entry applied to state machine
	nextIndex   []int // for each server, index of the next log entry to send
	matchIndex  []int // for each server, index of highest log entry known to be replicated
	applyCh     chan ApplyMsg

	lastIncludedIndex int    // Index of last entry in snapshot
	lastIncludedTerm  int    // Term of last entry in snapshot
	snapshot          []byte // The actual snapshot data
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (3A).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	isleader = rf.state == "leader"

	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).
func (rf *Raft) persist() {
	// Your code here (3C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)

	// rf.mu.Lock()
	// defer rf.mu.Unlock()

	// Create a buffer to hold serialized data
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)

	// Encode persistent state
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	e.Encode(rf.log)
	// e.Encode(rf.lastApplied) // Also persist lastApplied
	// e.Encode(rf.commitIndex) // and commitIndex
	e.Encode(rf.lastIncludedIndex) // NEW
	e.Encode(rf.lastIncludedTerm)  // NEW

	// Save serialized state
	raftstate := w.Bytes()
	rf.persister.Save(raftstate, rf.snapshot)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (3C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
	// Create a decoder to deserialize data
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)

	var currentTerm int
	var votedFor int
	var logEntries []LogEntry
	// var lastApplied int
	// var commitIndex int
	var lastIncludedIndex int
	var lastIncludedTerm int

	// Decode persistent state
	if d.Decode(&currentTerm) != nil ||
		d.Decode(&votedFor) != nil ||
		d.Decode(&logEntries) != nil ||
		d.Decode(&lastIncludedIndex) != nil ||
		d.Decode(&lastIncludedTerm) != nil {
		log.Fatalf("Failed to decode persisted state")
	} else {
		rf.currentTerm = currentTerm
		rf.votedFor = votedFor
		rf.log = logEntries
		// rf.lastApplied = lastApplied
		// rf.commitIndex = commitIndex
		rf.lastIncludedIndex = lastIncludedIndex
		rf.lastIncludedTerm = lastIncludedTerm
		// Restore snapshot
		rf.snapshot = rf.persister.ReadSnapshot()

		// Update lastApplied to at least the snapshot index
		if rf.lastApplied < rf.lastIncludedIndex {
			rf.lastApplied = rf.lastIncludedIndex
		}
	}
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// Don't snapshot if we've already included this index or beyond
	if index <= rf.lastIncludedIndex {
		return
	}

	// Don't snapshot beyond our current log
	if index > rf.getLastLogIndex() {
		return
	}

	// Find the term at the snapshot index
	snapshotTerm := rf.getLogTerm(index)
	if snapshotTerm == -1 {
		return // Invalid index
	}

	// Trim the log - keep only entries after the snapshot index
	newLogStartIndex := rf.logIndex(index) + 1
	if newLogStartIndex >= len(rf.log) {
		// Snapshot includes all current log entries
		rf.log = make([]LogEntry, 1) // Keep dummy entry at index 0
		rf.log[0] = LogEntry{Term: snapshotTerm, Command: nil}
	} else {
		// Keep entries after the snapshot index
		newLog := make([]LogEntry, 1) // Start with dummy entry
		newLog[0] = LogEntry{Term: snapshotTerm, Command: nil}
		newLog = append(newLog, rf.log[newLogStartIndex:]...)
		rf.log = newLog
	}

	// Update snapshot state
	rf.lastIncludedIndex = index
	rf.lastIncludedTerm = snapshotTerm
	rf.snapshot = make([]byte, len(snapshot))
	copy(rf.snapshot, snapshot)

	// Update lastApplied if necessary
	if rf.lastApplied < index {
		rf.lastApplied = index
	}

	// Persist the new state and snapshot
	rf.persist()
}

type InstallSnapshotArgs struct {
	Term              int
	LeaderId          int
	LastIncludedIndex int
	LastIncludedTerm  int
	Data              []byte
}

type InstallSnapshotReply struct {
	Term int
}

func (rf *Raft) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	DPrintf("Server %d received InstallSnapshot from %d, term %d, lastIncluded %d",
		rf.me, args.LeaderId, args.Term, args.LastIncludedIndex)

	reply.Term = rf.currentTerm

	if args.Term < rf.currentTerm {
		DPrintf("Server %d rejecting InstallSnapshot due to lower term %d < %d",
			rf.me, args.Term, rf.currentTerm)
		return
	}

	if args.Term > rf.currentTerm {
		DPrintf("Server %d updating term from %d to %d", rf.me, rf.currentTerm, args.Term)
		rf.becomeFollower(args.Term)
		reply.Term = rf.currentTerm
	}

	rf.electionTimer.Reset(getRandomElectionTimeout())

	if args.LastIncludedIndex <= rf.lastIncludedIndex {
		DPrintf("Server %d ignoring old snapshot %d <= %d",
			rf.me, args.LastIncludedIndex, rf.lastIncludedIndex)
		return
	}

	DPrintf("Server %d applying snapshot up to index %d", rf.me, args.LastIncludedIndex)

	// Save snapshot
	rf.snapshot = make([]byte, len(args.Data))
	copy(rf.snapshot, args.Data)

	// Trim log
	rf.log = make([]LogEntry, 1)
	DPrintf("rf.log at creation is %d ", rf.log)
	rf.log[0] = LogEntry{Term: args.LastIncludedTerm, Command: nil}
	DPrintf("rf.log at after first entry is %d ", rf.log)

	rf.lastIncludedIndex = args.LastIncludedIndex
	rf.lastIncludedTerm = args.LastIncludedTerm

	if rf.lastApplied < args.LastIncludedIndex {
		rf.lastApplied = args.LastIncludedIndex
	}
	if rf.commitIndex < args.LastIncludedIndex {
		rf.commitIndex = args.LastIncludedIndex
	}

	rf.persist()

	// Send snapshot asynchronously to avoid deadlock
	snapshotData := make([]byte, len(args.Data))
	copy(snapshotData, args.Data)

	applyMsg := ApplyMsg{
		SnapshotValid: true,
		Snapshot:      snapshotData,
		SnapshotTerm:  args.LastIncludedTerm,
		SnapshotIndex: args.LastIncludedIndex,
	}

	// Send in a separate goroutine to avoid blocking while holding the lock
	go func() {
		rf.applyCh <- applyMsg
	}()
	DPrintf("Server %d: InstallSnapshot complete, lastIncluded now %d", rf.me, rf.lastIncludedIndex)

}

func (rf *Raft) sendInstallSnapshot(server int, args *InstallSnapshotArgs, reply *InstallSnapshotReply) bool {
	ok := rf.peers[server].Call("Raft.InstallSnapshot", args, reply)
	return ok
}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Your data here (3A, 3B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (3A).
	Term        int
	VoteGranted bool
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}

	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.state = "follower"
		rf.votedFor = -1
		rf.persist()
	}

	reply.Term = rf.currentTerm
	if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) &&
		rf.isLogUpToDate(args.LastLogIndex, args.LastLogTerm) {
		rf.votedFor = args.CandidateId
		reply.VoteGranted = true
		rf.electionTimer.Reset(getRandomElectionTimeout())
		rf.persist()
	} else {
		reply.VoteGranted = false
	}
}

func (rf *Raft) isLogUpToDate(candidateLastLogIndex int, candidateLastLogTerm int) bool {
	myLastLogIndex := rf.getLastLogIndex()
	myLastLogTerm := rf.getLastLogTerm()

	if candidateLastLogTerm > myLastLogTerm {
		return true
	}
	if candidateLastLogTerm == myLastLogTerm && candidateLastLogIndex >= myLastLogIndex {
		return true
	}
	return false
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
	XTerm   int // Term of conflicting entry (if any)
	XIndex  int // Index of first entry with XTerm (if any)
	XLen    int // Length of follower's log
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Success = false
	reply.Term = rf.currentTerm
	reply.XTerm = -1
	reply.XIndex = -1
	reply.XLen = rf.getLastLogIndex() + 1

	if args.Term < rf.currentTerm {
		return
	}

	if args.Term > rf.currentTerm {
		rf.becomeFollower(args.Term)
		reply.Term = rf.currentTerm
	}

	rf.electionTimer.Reset(getRandomElectionTimeout())

	// Check if prevLogIndex is in our snapshot
	if args.PrevLogIndex < rf.lastIncludedIndex {
		// Leader is behind our snapshot
		return
	}

	// Check if we have the previous log entry
	if args.PrevLogIndex > rf.getLastLogIndex() {
		// We don't have enough entries
		reply.XLen = rf.getLastLogIndex() + 1
		return
	}

	// Check if the previous log entry matches
	if args.PrevLogIndex > rf.lastIncludedIndex {
		arrayIndex := rf.logIndex(args.PrevLogIndex)
		if arrayIndex < 0 || arrayIndex >= len(rf.log) || rf.log[arrayIndex].Term != args.PrevLogTerm {
			// Log inconsistency
			if arrayIndex >= 0 && arrayIndex < len(rf.log) {
				reply.XTerm = rf.log[arrayIndex].Term
				// Find first index with this term
				for i := arrayIndex; i >= 0; i-- {
					if rf.log[i].Term != reply.XTerm {
						reply.XIndex = rf.realIndex(i + 1)
						break
					}
				}
				if reply.XIndex == -1 {
					reply.XIndex = rf.lastIncludedIndex + 1
				}
			}
			return
		}
	} else if args.PrevLogIndex == rf.lastIncludedIndex {
		// Previous entry is the last entry in snapshot
		if args.PrevLogTerm != rf.lastIncludedTerm {
			return
		}
	}

	reply.Success = true

	// Append new entries
	for i, entry := range args.Entries {
		realIndex := args.PrevLogIndex + 1 + i

		if realIndex <= rf.lastIncludedIndex {
			// Skip entries that are already in snapshot
			continue
		}

		arrayIndex := rf.logIndex(realIndex)

		if arrayIndex < len(rf.log) {
			if rf.log[arrayIndex].Term != entry.Term {
				// Delete conflicting entries and all that follow
				rf.log = rf.log[:arrayIndex]
				rf.log = append(rf.log, entry)
			}
		} else {
			// Append new entry
			rf.log = append(rf.log, entry)
		}
	}

	// Update commit index
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = min(args.LeaderCommit, rf.getLastLogIndex())
	}

	rf.persist()
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.state != "leader" {
		return -1, rf.currentTerm, false
	}

	if command == nil {
		return -1, rf.currentTerm, false
	}
	DPrintf("Server %d starting agreement for command, term %d", rf.me, rf.currentTerm)

	// Append the command to the leader's log
	entry := LogEntry{
		Term:    rf.currentTerm,
		Command: command,
	}

	rf.log = append(rf.log, entry)
	rf.persist()

	// Get the real index of the new entry
	realIndex := rf.getLastLogIndex()

	// Update leader's matchIndex for itself
	rf.matchIndex[rf.me] = realIndex

	DPrintf("Server %d appended entry at index %d, starting replication", rf.me, realIndex)

	// Only start replication occasionally to reduce congestion
	// Let heartbeats handle most replication
	if realIndex%5 == 0 { // Only replicate every 5th command immediately
		for i := range rf.peers {
			if i != rf.me {
				go func(peer int) {
					rf.replicateOneRound(peer)
				}(i)
			}
		}
	}

	return realIndex, rf.currentTerm, true
}

func (rf *Raft) replicateOneRound(peer int) {
	rf.mu.Lock()
	if rf.state != "leader" {
		rf.mu.Unlock()
		return
	}

	nextIdx := rf.nextIndex[peer]
	currentTerm := rf.currentTerm
	leaderId := rf.me

	// Check if we need to send a snapshot
	if nextIdx <= rf.lastIncludedIndex {
		// Send InstallSnapshot RPC
		args := &InstallSnapshotArgs{
			Term:              currentTerm,
			LeaderId:          leaderId,
			LastIncludedIndex: rf.lastIncludedIndex,
			LastIncludedTerm:  rf.lastIncludedTerm,
			Data:              make([]byte, len(rf.snapshot)),
		}
		copy(args.Data, rf.snapshot)
		rf.mu.Unlock()

		reply := &InstallSnapshotReply{}
		if rf.sendInstallSnapshot(peer, args, reply) {
			rf.mu.Lock()
			// Validate we're still leader with same term
			if rf.state == "leader" && rf.currentTerm == currentTerm {
				if reply.Term > rf.currentTerm {
					rf.becomeFollower(reply.Term)
				} else {
					// Update nextIndex and matchIndex on success
					rf.nextIndex[peer] = rf.lastIncludedIndex + 1
					rf.matchIndex[peer] = rf.lastIncludedIndex
				}
			}
			rf.mu.Unlock()
		}
		return
	}

	// Skip if already up to date
	if nextIdx > rf.getLastLogIndex() {
		rf.mu.Unlock()
		return
	}

	// Prepare AppendEntries
	prevLogIndex := nextIdx - 1
	prevLogTerm := rf.getLogTerm(prevLogIndex)
	if prevLogTerm == -1 {
		rf.mu.Unlock()
		return
	}

	var entries []LogEntry
	if nextIdx <= rf.getLastLogIndex() {
		startArrayIndex := rf.logIndex(nextIdx)
		if startArrayIndex >= 0 && startArrayIndex < len(rf.log) {
			entries = make([]LogEntry, len(rf.log)-startArrayIndex)
			copy(entries, rf.log[startArrayIndex:])
		}
	}

	args := &AppendEntriesArgs{
		Term:         currentTerm,
		LeaderId:     leaderId,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: rf.commitIndex,
	}
	rf.mu.Unlock()

	reply := &AppendEntriesReply{}
	if rf.sendAppendEntries(peer, args, reply) {
		rf.mu.Lock()
		// Validate we're still leader with same term
		if rf.state == "leader" && rf.currentTerm == currentTerm {
			if reply.Term > rf.currentTerm {
				rf.becomeFollower(reply.Term)
			} else if reply.Success {
				// Update indices on success
				matchIndex := prevLogIndex + len(entries)
				if matchIndex > rf.matchIndex[peer] {
					rf.matchIndex[peer] = matchIndex
					rf.nextIndex[peer] = matchIndex + 1
					rf.updateCommitIndex()
				}
			} else {
				// Handle log inconsistency
				if reply.XTerm == -1 {
					rf.nextIndex[peer] = reply.XLen
				} else {
					lastIndex := -1
					for i := rf.getLastLogIndex(); i > rf.lastIncludedIndex; i-- {
						if rf.getLogTerm(i) == reply.XTerm {
							lastIndex = i
							break
						}
					}
					if lastIndex != -1 {
						rf.nextIndex[peer] = lastIndex + 1
					} else {
						rf.nextIndex[peer] = reply.XIndex
					}
				}
				// Ensure nextIndex doesn't go below lastIncludedIndex + 1
				if rf.nextIndex[peer] <= rf.lastIncludedIndex {
					rf.nextIndex[peer] = rf.lastIncludedIndex + 1
				}
			}
		}
		rf.mu.Unlock()
	}
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) ticker() {
	for !rf.killed() {
		select {
		case <-rf.heartbeatTimer.C:
			if rf.state == "leader" {
				rf.sendHeartbeats()
			}
			rf.heartbeatTimer.Reset(100 * time.Millisecond)
		case <-rf.electionTimer.C:
			if rf.state != "leader" {
				rf.startElection()
			}
			rf.electionTimer.Reset(getRandomElectionTimeout())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (rf *Raft) becomeLeader() {
	rf.state = "leader"

	// Initialize nextIndex and matchIndex for all peers
	lastLogIndex := rf.getLastLogIndex()
	for i := range rf.peers {
		rf.nextIndex[i] = lastLogIndex + 1
		rf.matchIndex[i] = 0
	}

	// Set our own matchIndex to our last log index
	rf.matchIndex[rf.me] = lastLogIndex

	rf.persist()

	// Send initial heartbeats
	rf.heartbeatTimer.Reset(0)
}

func (rf *Raft) startElection() {
	rf.mu.Lock()
	rf.state = "candidate"
	rf.currentTerm++
	rf.votedFor = rf.me
	rf.electionTimer.Reset(getRandomElectionTimeout())

	currentTerm := rf.currentTerm
	lastLogIndex := rf.getLastLogIndex()
	lastLogTerm := rf.getLastLogTerm()
	rf.persist()
	rf.mu.Unlock()

	votes := int32(1) // Vote for ourselves

	for i := range rf.peers {
		if i != rf.me {
			go func(peer int) {
				args := &RequestVoteArgs{
					Term:         currentTerm,
					CandidateId:  rf.me,
					LastLogIndex: lastLogIndex,
					LastLogTerm:  lastLogTerm,
				}

				reply := &RequestVoteReply{}
				if rf.sendRequestVote(peer, args, reply) {
					rf.mu.Lock()

					if rf.state != "candidate" || rf.currentTerm != currentTerm {
						rf.mu.Unlock()
						return
					}

					if reply.Term > rf.currentTerm {
						rf.becomeFollower(reply.Term)
						rf.mu.Unlock()
						return
					}

					if reply.VoteGranted {
						newVotes := atomic.AddInt32(&votes, 1)
						if int(newVotes) > len(rf.peers)/2 && rf.state == "candidate" {
							rf.becomeLeader()
						}
					}

					rf.mu.Unlock()
				}
			}(i)
		}
	}
}

func (rf *Raft) sendHeartbeats() {
	rf.mu.Lock()
	if rf.state != "leader" {
		rf.mu.Unlock()
		return
	}

	term := rf.currentTerm
	rf.mu.Unlock()

	for i := range rf.peers {
		if i != rf.me {
			go func(peer int) {
				rf.mu.Lock()
				if rf.state != "leader" || rf.currentTerm != term {
					rf.mu.Unlock()
					return
				}

				// Get the next index to send to this peer (real index)
				nextIdx := rf.nextIndex[peer]

				// Check if we need to send a snapshot
				if nextIdx <= rf.lastIncludedIndex {
					// Send InstallSnapshot RPC
					args := &InstallSnapshotArgs{
						Term:              term,
						LeaderId:          rf.me,
						LastIncludedIndex: rf.lastIncludedIndex,
						LastIncludedTerm:  rf.lastIncludedTerm,
						Data:              make([]byte, len(rf.snapshot)),
					}
					copy(args.Data, rf.snapshot)

					rf.mu.Unlock()

					var reply InstallSnapshotReply
					if rf.sendInstallSnapshot(peer, args, &reply) {
						rf.mu.Lock()
						defer rf.mu.Unlock()
						if rf.state != "leader" || rf.currentTerm != term {
							return
						}

						if reply.Term > rf.currentTerm {
							rf.becomeFollower(reply.Term)
							return
						}

						// Update nextIndex and matchIndex on success
						rf.nextIndex[peer] = rf.lastIncludedIndex + 1
						rf.matchIndex[peer] = rf.lastIncludedIndex
					} else {
						// Add backoff for failed snapshot installations
						time.Sleep(10 * time.Millisecond)
					}
					return
				}

				// Calculate prevLogIndex (real index)
				prevLogIndex := nextIdx - 1
				prevLogTerm := 0

				if prevLogIndex == rf.lastIncludedIndex {
					// Previous entry is the last entry in snapshot
					prevLogTerm = rf.lastIncludedTerm
				} else if prevLogIndex > rf.lastIncludedIndex {
					// Previous entry is in our current log
					arrayIndex := rf.logIndex(prevLogIndex)
					if arrayIndex >= 0 && arrayIndex < len(rf.log) {
						prevLogTerm = rf.log[arrayIndex].Term
					} else {
						// Invalid index, skip this peer
						rf.mu.Unlock()
						return
					}
				}

				// Prepare entries to send
				var entries []LogEntry
				if nextIdx <= rf.getLastLogIndex() {
					startArrayIndex := rf.logIndex(nextIdx)
					if startArrayIndex >= 0 && startArrayIndex < len(rf.log) {
						entries = make([]LogEntry, len(rf.log)-startArrayIndex)
						copy(entries, rf.log[startArrayIndex:])
					}
				}

				args := &AppendEntriesArgs{
					Term:         term,
					LeaderId:     rf.me,
					PrevLogIndex: prevLogIndex,
					PrevLogTerm:  prevLogTerm,
					Entries:      entries,
					LeaderCommit: rf.commitIndex,
				}

				rf.mu.Unlock()

				var reply AppendEntriesReply
				if rf.sendAppendEntries(peer, args, &reply) {
					rf.mu.Lock()
					defer rf.mu.Unlock()

					if rf.state != "leader" || rf.currentTerm != term {
						return
					}

					if reply.Term > rf.currentTerm {
						rf.becomeFollower(reply.Term)
						return
					}

					if reply.Success {
						// Update indices on success
						newMatchIndex := prevLogIndex + len(entries)
						if newMatchIndex > rf.matchIndex[peer] {
							rf.matchIndex[peer] = newMatchIndex
							rf.nextIndex[peer] = newMatchIndex + 1
							rf.updateCommitIndex()
						}
					} else {
						// Handle log inconsistency
						if reply.XTerm == -1 {
							// Case 3: Follower's log is too short
							rf.nextIndex[peer] = reply.XLen
						} else {
							// Find the last entry with XTerm in leader's log
							lastIndex := -1
							for i := rf.getLastLogIndex(); i > rf.lastIncludedIndex; i-- {
								if rf.getLogTerm(i) == reply.XTerm {
									lastIndex = i
									break
								}
							}

							if lastIndex != -1 {
								// Case 2: Leader has XTerm
								rf.nextIndex[peer] = lastIndex + 1
							} else {
								// Case 1: Leader doesn't have XTerm
								rf.nextIndex[peer] = reply.XIndex
							}
						}

						// Ensure nextIndex doesn't go below lastIncludedIndex + 1
						if rf.nextIndex[peer] <= rf.lastIncludedIndex {
							rf.nextIndex[peer] = rf.lastIncludedIndex + 1
						}
					}
				}
			}(i)
		}
	}
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (3A, 3B, 3C).
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.log = make([]LogEntry, 1)
	// rf.log[0] = LogEntry{Term: 0} // Dummy entry to match paper's 1-indexing
	rf.state = "follower"

	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))
	rf.applyCh = applyCh

	rf.lastIncludedIndex = 0
	rf.lastIncludedTerm = 0
	rf.snapshot = nil

	rf.electionTimer = time.NewTimer(getRandomElectionTimeout())
	rf.heartbeatTimer = time.NewTimer(100 * time.Millisecond)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// If we have a snapshot, apply it to the service
	if rf.snapshot != nil && len(rf.snapshot) > 0 {
		go func() {
			rf.applyCh <- ApplyMsg{
				SnapshotValid: true,
				Snapshot:      rf.snapshot,
				SnapshotTerm:  rf.lastIncludedTerm,
				SnapshotIndex: rf.lastIncludedIndex,
			}
		}()
	}

	// start ticker goroutine to start elections
	go rf.ticker()

	// Add to Make() function
	go func() {
		for !rf.killed() {
			time.Sleep(10 * time.Millisecond)
			rf.mu.Lock()
			if rf.commitIndex > rf.lastApplied {
				rf.mu.Unlock()
				rf.applyCommittedEntries()
			} else {
				rf.mu.Unlock()
			}
		}
	}()

	return rf
}

func getRandomElectionTimeout() time.Duration {
	return time.Duration(500+rand.Intn(250)) * time.Millisecond
}

func (rf *Raft) updateCommitIndex() {
	// For each log entry starting from the next uncommitted real index
	for realIndex := rf.commitIndex + 1; realIndex <= rf.getLastLogIndex(); realIndex++ {
		// Count how many servers have this entry
		count := 1 // Count ourselves
		for peer := range rf.peers {
			if peer != rf.me && rf.matchIndex[peer] >= realIndex {
				count++
			}
		}

		// Get the array index for this real index
		arrayIndex := rf.logIndex(realIndex)
		if arrayIndex < 0 || arrayIndex >= len(rf.log) {
			break // Invalid index
		}

		// If majority and from current term, commit it
		if count > len(rf.peers)/2 && rf.log[arrayIndex].Term == rf.currentTerm {
			rf.commitIndex = realIndex
		} else {
			// No majority, stop checking further entries
			break
		}
	}
}

func (rf *Raft) applyCommittedEntries() {
	rf.mu.Lock()
	if rf.lastApplied >= rf.commitIndex {
		rf.mu.Unlock()
		return
	}

	// Create a batch of messages to apply
	var messages []ApplyMsg
	for realIndex := rf.lastApplied + 1; realIndex <= rf.commitIndex; realIndex++ {
		// Skip entries that are in the snapshot
		if realIndex <= rf.lastIncludedIndex {
			continue
		}

		arrayIndex := rf.logIndex(realIndex)
		if arrayIndex < 0 || arrayIndex >= len(rf.log) {
			break // Invalid index
		}

		messages = append(messages, ApplyMsg{
			CommandValid: true,
			Command:      rf.log[arrayIndex].Command,
			CommandIndex: realIndex,
		})
	}

	// Update lastApplied before releasing lock
	if len(messages) > 0 {
		rf.lastApplied = messages[len(messages)-1].CommandIndex
	}

	rf.mu.Unlock()

	// Apply messages without holding the lock
	// Apply messages without holding the lock, with timeout
	for _, msg := range messages {
		select {
		case rf.applyCh <- msg:
			// Successfully sent
		case <-time.After(50 * time.Millisecond):
			// Timeout - send in background to avoid blocking
			DPrintf("Apply channel blocked for server %d, msg index %d", rf.me, msg.CommandIndex)
			go func(message ApplyMsg) {
				rf.applyCh <- message
			}(msg)
		}
	}

}

func (rf *Raft) becomeFollower(term int) {
	rf.state = "follower"
	rf.currentTerm = term
	rf.votedFor = -1
	rf.persist()
}

// Convert real log index to array index in our trimmed log
func (rf *Raft) logIndex(realIndex int) int {
	if realIndex <= rf.lastIncludedIndex {
		return -1 // Invalid
	}
	return realIndex - rf.lastIncludedIndex
}

// Convert array index to real log index
func (rf *Raft) realIndex(logIndex int) int {
	return logIndex + rf.lastIncludedIndex
}

// Get the last log index (real index)
func (rf *Raft) getLastLogIndex() int {
	return rf.lastIncludedIndex + len(rf.log) - 1
}

// Get the last log term
func (rf *Raft) getLastLogTerm() int {
	if len(rf.log) == 0 {
		return rf.lastIncludedTerm
	}
	return rf.log[len(rf.log)-1].Term
}

// Get term at a specific real index
func (rf *Raft) getLogTerm(realIndex int) int {
	if realIndex == rf.lastIncludedIndex {
		return rf.lastIncludedTerm
	}
	if realIndex < rf.lastIncludedIndex {
		// This shouldn't happen in normal operation
		return -1
	}
	arrayIndex := rf.logIndex(realIndex)
	if arrayIndex >= len(rf.log) {
		return -1
	}
	return rf.log[arrayIndex].Term
}

func (rf *Raft) hasLogWithTerm(term int) bool {
	for _, entry := range rf.log {
		if entry.Term == term {
			return true
		}
	}
	return false
}

func (rf *Raft) findLastIndexWithTerm(term int) int {
	for i := len(rf.log) - 1; i >= 0; i-- {
		if rf.log[i].Term == term {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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
	//	"bytes"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"github.com/paras-bhavnani/distributed-kv-store-raft/labgob"
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
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (3D).

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
	// Your code here (3A, 3B).
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
	}

	reply.Term = rf.currentTerm

	if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) &&
		(len(rf.log) == 0 || args.LastLogTerm > rf.log[len(rf.log)-1].Term ||
			(args.LastLogTerm == rf.log[len(rf.log)-1].Term && args.LastLogIndex >= len(rf.log)-1)) {
		rf.votedFor = args.CandidateId
		reply.VoteGranted = true
		rf.electionTimer.Reset(getRandomElectionTimeout())
	} else {
		reply.VoteGranted = false
	}
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
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term = rf.currentTerm
	reply.Success = false

	if args.Term < rf.currentTerm {
		return
	}

	rf.electionTimer.Reset(getRandomElectionTimeout())

	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.state = "follower"
		rf.votedFor = -1
		rf.persist()
	}

	// Implement log consistency check and append new entries
	// For 3A, we only need to handle heartbeats (empty Entries)
	// Check log consistency
	if args.PrevLogIndex >= len(rf.log) {
		return
	}
	if args.PrevLogIndex >= 0 && rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		return
	}

	// Append new entries, truncating any conflicting entries
	if len(args.Entries) > 0 {
		rf.log = rf.log[:args.PrevLogIndex+1]
		rf.log = append(rf.log, args.Entries...)
		rf.persist()
	}

	// Update commit index and apply entries
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = min(args.LeaderCommit, len(rf.log)-1)
		go rf.applyCommitted()
	}

	reply.Success = true
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
		return -1, -1, false
	}

	// Append the command to the leader's log
	index := len(rf.log) // Current length will be the new index
	entry := LogEntry{
		Term:    rf.currentTerm,
		Command: command,
	}
	rf.log = append(rf.log, entry)
	rf.persist()

	// Update leader's matchIndex for itself
	rf.matchIndex[rf.me] = index

	// Start replication immediately
	for i := range rf.peers {
		if i != rf.me {
			go func(peer int, currentTerm int) {
				rf.replicateOneRound(peer)
			}(i, rf.currentTerm)
		}
	}

	DPrintf("Leader %d: Starting command at index %d", rf.me, index)
	return index, rf.currentTerm, true
}

func (rf *Raft) replicateOneRound(peer int) {
	for !rf.killed() {
		rf.mu.Lock()
		if rf.state != "leader" {
			rf.mu.Unlock()
			return
		}

		// If already up to date, no need to send anything
		if rf.nextIndex[peer] >= len(rf.log) {
			rf.mu.Unlock()
			return
		}

		prevLogIndex := rf.nextIndex[peer] - 1
		prevLogTerm := 0
		if prevLogIndex >= 0 && prevLogIndex < len(rf.log) {
			prevLogTerm = rf.log[prevLogIndex].Term
		}

		// Create a copy of entries to send
		entries := make([]LogEntry, len(rf.log[rf.nextIndex[peer]:]))
		copy(entries, rf.log[rf.nextIndex[peer]:])

		args := &AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderId:     rf.me,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      entries,
			LeaderCommit: rf.commitIndex,
		}
		currentTerm := rf.currentTerm
		rf.mu.Unlock()

		reply := &AppendEntriesReply{}
		if rf.sendAppendEntries(peer, args, reply) {
			rf.mu.Lock()
			if rf.state != "leader" || rf.currentTerm != currentTerm {
				rf.mu.Unlock()
				return
			}

			if reply.Success {
				// Update matchIndex and nextIndex
				newMatchIndex := prevLogIndex + len(entries)
				rf.matchIndex[peer] = newMatchIndex
				rf.nextIndex[peer] = newMatchIndex + 1

				// Try to commit new entries
				for n := len(rf.log) - 1; n > rf.commitIndex; n-- {
					if rf.log[n].Term == rf.currentTerm {
						count := 1 // count ourselves
						for p := range rf.peers {
							if p != rf.me && rf.matchIndex[p] >= n {
								count++
							}
						}
						if count > len(rf.peers)/2 {
							rf.commitIndex = n
							go rf.applyCommitted()
							break
						}
					}
				}
				rf.mu.Unlock()
				return
			} else {
				if reply.Term > rf.currentTerm {
					rf.currentTerm = reply.Term
					rf.state = "follower"
					rf.votedFor = -1
					rf.persist()
					rf.mu.Unlock()
					return
				}
				// Decrement nextIndex and retry
				if rf.nextIndex[peer] > 1 {
					rf.nextIndex[peer]--
				}
				rf.mu.Unlock()
				continue
			}
		} else {
			return
		}
	}
}

func (rf *Raft) applyCommitted() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	for rf.lastApplied < rf.commitIndex {
		rf.lastApplied++
		if rf.lastApplied < len(rf.log) {
			msg := ApplyMsg{
				CommandValid: true,
				Command:      rf.log[rf.lastApplied].Command,
				CommandIndex: rf.lastApplied,
			}
			rf.applyCh <- msg
		}
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
	for rf.killed() == false {

		// Your code here (3A)
		// Check if a leader election should be started.
		select {
		case <-rf.heartbeatTimer.C:
			if rf.state == "leader" {
				rf.sendHeartbeats()
			}
		case <-rf.electionTimer.C:
			if rf.state != "leader" {
				rf.startElection()
			}
		}

		// pause for a random amount of time between 50 and 350
		// milliseconds.
		ms := 50 + (rand.Int63() % 300)
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

func (rf *Raft) becomeLeader() {
	rf.state = "leader"
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))

	for i := range rf.peers {
		rf.nextIndex[i] = len(rf.log)
		rf.matchIndex[i] = 0
	}

	rf.sendHeartbeats()
}

func (rf *Raft) startElection() {
	rf.mu.Lock()
	rf.currentTerm++
	rf.state = "candidate"
	rf.votedFor = rf.me
	term := rf.currentTerm
	rf.mu.Unlock()

	votes := 1
	for i := range rf.peers {
		if i != rf.me {
			go func(peer int) {
				args := &RequestVoteArgs{
					Term:         term,
					CandidateId:  rf.me,
					LastLogIndex: len(rf.log) - 1,
					LastLogTerm:  rf.getLastLogTerm(),
				}
				reply := &RequestVoteReply{}
				if rf.sendRequestVote(peer, args, reply) {
					rf.mu.Lock()
					defer rf.mu.Unlock()
					if reply.VoteGranted && rf.state == "candidate" && rf.currentTerm == term {
						votes++
						if votes > len(rf.peers)/2 {
							rf.becomeLeader()
						}
					} else if reply.Term > rf.currentTerm {
						rf.currentTerm = reply.Term
						rf.state = "follower"
						rf.votedFor = -1
					}
				}
			}(i)
		}
	}

	rf.electionTimer.Reset(getRandomElectionTimeout())
}

func (rf *Raft) sendHeartbeats() {
	for i := range rf.peers {
		if i != rf.me {
			go func(peer int) {
				rf.mu.Lock()
				if rf.state != "leader" {
					rf.mu.Unlock()
					return
				}

				prevLogIndex := rf.nextIndex[peer] - 1
				prevLogTerm := 0
				if prevLogIndex >= 0 && prevLogIndex < len(rf.log) {
					prevLogTerm = rf.log[prevLogIndex].Term
				}

				// Only send entries for heartbeats if needed
				var entries []LogEntry
				if rf.nextIndex[peer] < len(rf.log) {
					entries = make([]LogEntry, len(rf.log)-rf.nextIndex[peer])
					copy(entries, rf.log[rf.nextIndex[peer]:])
				}

				args := &AppendEntriesArgs{
					Term:         rf.currentTerm,
					LeaderId:     rf.me,
					PrevLogIndex: prevLogIndex,
					PrevLogTerm:  prevLogTerm,
					Entries:      entries,
					LeaderCommit: rf.commitIndex,
				}
				rf.mu.Unlock()

				reply := &AppendEntriesReply{}
				if rf.sendAppendEntries(peer, args, reply) {
					rf.mu.Lock()
					defer rf.mu.Unlock()

					if rf.state != "leader" || rf.currentTerm != args.Term {
						return
					}

					if reply.Success {
						rf.nextIndex[peer] = prevLogIndex + len(entries) + 1
						rf.matchIndex[peer] = rf.nextIndex[peer] - 1

						// Check if we can commit any new entries
						for N := len(rf.log) - 1; N > rf.commitIndex; N-- {
							if rf.log[N].Term == rf.currentTerm {
								count := 1 // count ourselves
								for p := range rf.peers {
									if p != rf.me && rf.matchIndex[p] >= N {
										count++
									}
								}
								if count > len(rf.peers)/2 {
									rf.commitIndex = N
									break
								}
							}
						}
					} else {
						rf.nextIndex[peer]--
					}
				}
			}(i)
		}
	}
	rf.heartbeatTimer.Reset(100 * time.Millisecond)
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
	rf.state = "follower"

	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))
	rf.applyCh = applyCh

	rf.electionTimer = time.NewTimer(getRandomElectionTimeout())
	rf.heartbeatTimer = time.NewTimer(100 * time.Millisecond)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()

	// Add to Make() function
	go func() {
		for !rf.killed() {
			time.Sleep(10 * time.Millisecond)
			rf.applyCommittedEntries()
		}
	}()

	return rf
}

func getRandomElectionTimeout() time.Duration {
	return time.Duration(500+rand.Intn(250)) * time.Millisecond
}

func (rf *Raft) getLastLogTerm() int {
	if len(rf.log) > 0 {
		return rf.log[len(rf.log)-1].Term
	}
	return 0
}

func (rf *Raft) applyCommittedEntries() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	for rf.lastApplied < rf.commitIndex {
		rf.lastApplied++
		applyMsg := ApplyMsg{
			CommandValid: true,
			Command:      rf.log[rf.lastApplied].Command,
			CommandIndex: rf.lastApplied,
		}
		rf.applyCh <- applyMsg
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

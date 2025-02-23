package kvsrv

type Args interface {
	getClerkId() int64
	getOpSeq() int64
}

// Put or Append
type PutAppendArgs struct {
	Key   string
	Value string
	// You'll have to add definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	ClerkId int64
	OpSeq   int64
}

func (a *PutAppendArgs) getClerkId() int64 {
	return a.ClerkId
}

func (a *PutAppendArgs) getOpSeq() int64 {
	return a.OpSeq
}

type PutAppendReply struct {
	Value string
}

type GetArgs struct {
	Key string
	// You'll have to add definitions here.
	ClerkId int64
	OpSeq   int64
}

func (a *GetArgs) getClerkId() int64 {
	return a.ClerkId
}
func (a *GetArgs) getOpSeq() int64 {
	return a.OpSeq
}

type GetReply struct {
	Value string
}

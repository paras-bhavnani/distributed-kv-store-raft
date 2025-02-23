package kvsrv

import (
	"fmt"
	"log"
	"sync"
)

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type KVServer struct {
	mu sync.Mutex

	db        map[string]string
	opResults map[int64]string
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	reply.Value = kv.db[args.Key]
}

func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.db[args.Key] = args.Value
	reply.Value = kv.db[args.Key]
}

func (kv *KVServer) Append(args *PutAppendArgs, reply *PutAppendReply) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	if kv.isDuplicated(args) {
		reply.Value = kv.opResults[args.getOpSeq()] // reply cached result
		fmt.Println("Duplicated")
		fmt.Println(reply.Value)
		return
	}
	kv.opResults[args.getOpSeq()] = kv.db[args.Key] // to cache result of this operation
	reply.Value = kv.db[args.Key]
	kv.db[args.Key] += args.Value
}

func StartKVServer() *KVServer {
	kv := new(KVServer)

	kv.db = make(map[string]string)
	kv.opResults = make(map[int64]string)

	return kv
}

func (kv *KVServer) isDuplicated(arg Args) bool {
	_, exist := kv.opResults[arg.getOpSeq()]
	return exist
}

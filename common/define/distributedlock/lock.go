package distributedlock

type DistributedLock interface {
	Lock()error
	Unlock()error
}
package eventbus

import "sync"

var (
	runtimeMu     sync.RWMutex
	runtimeQueues QueueManager
	dispatchWake  = make(chan struct{}, 1)
)

func SetQueueManager(manager QueueManager) {
	runtimeMu.Lock()
	runtimeQueues = manager
	runtimeMu.Unlock()
}

func SignalDispatch() {
	select {
	case dispatchWake <- struct{}{}:
	default:
	}
}

func DispatchWake() <-chan struct{} { return dispatchWake }

func QueueManagerInstance() QueueManager {
	runtimeMu.RLock()
	defer runtimeMu.RUnlock()
	return runtimeQueues
}

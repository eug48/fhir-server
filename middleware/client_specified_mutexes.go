package client_specified_mutexes

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type LockRequest struct {
	mutexName string
	gateChannel chan int
}
type UnlockRequest struct {
	mutexName string
	lockId int
}

func Middleware() gin.HandlerFunc {
	var lockRequests = make (chan *LockRequest)
	var unlockRequests = make (chan *UnlockRequest)

	go func() {
		mutexes := make(map[string]map[int]chan int)

		for {
			select {
			case unlockRequest := <- unlockRequests:
				locks, present := mutexes[unlockRequest.mutexName]
				if !present {
					panic("client_specified_mutexes.go: mutex not present during unlock request")
				}
				_, lockIdPresent := locks[unlockRequest.lockId]
				if !lockIdPresent {
					panic("client_specified_mutexes.go: lockId not present during unlock request")
				}

				delete(locks, unlockRequest.lockId)
				if len(locks) > 0 {
					// pick an arbitrary waiter and tell them to proceed
					for lockId, gateChannel := range locks {
						fmt.Printf("[client_specified_mutexes] unlocked & releasing %d: %s\n", lockId, unlockRequest.mutexName)
						gateChannel <- lockId
						break
					}
				} else {
					fmt.Printf("[client_specified_mutexes] unlocked & freed: %s\n", unlockRequest.mutexName)
					delete (mutexes, unlockRequest.mutexName)
				}

			case lockRequest := <- lockRequests:
				locks, present := mutexes[lockRequest.mutexName]
				if present {
					// add to 'queue'
					newLockId := len(locks) + 1 // TODO: should worry about overflow? will very likely never have very many simultaneous requests on the same mutex..
					locks[newLockId] = lockRequest.gateChannel
					fmt.Printf("[client_specified_mutexes] queued %d: %s\n", newLockId, lockRequest.mutexName)
				} else {
					// save to a new queue and proceed
					locks := make(map[int]chan int)
					locks[0] = lockRequest.gateChannel
					lockRequest.gateChannel <- 0
					mutexes[lockRequest.mutexName] = locks
					fmt.Printf("[client_specified_mutexes] locked %d: %s\n", 0, lockRequest.mutexName)
				}

			}
		}

	}()

	return func(c *gin.Context) {

		mutexName := c.GetHeader("X-Mutex-Name")
		if mutexName != "" {
			lockRequest := &LockRequest{mutexName: mutexName, gateChannel: make(chan int) }
			lockRequests <- lockRequest
			lockId := <- lockRequest.gateChannel
			defer func() {
				unlockRequest := &UnlockRequest{mutexName, lockId }
				unlockRequests <- unlockRequest
			}()
			c.Header("X-Mutex-Used", "1")
		} else {
			c.Header("X-Mutex-Used", "0")
		}

		c.Next()
	}
}

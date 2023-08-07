//////////////////////////////////////////////////////////////////////
//
// Your video processing service has a freemium model. Everyone has 10
// sec of free processing time on your service. After that, the
// service will kill your process, unless you are a paid premium user.
//
// Beginner Level: 10s max per request
// Advanced Level: 10s max per user (accumulated)
//

package main

import (
	"log"
	"sync/atomic"
	"time"
)

const MAX_TIME = 10

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  int64 // in seconds
}

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func HandleRequest(process func(), u *User) bool {
	if u.IsPremium {
		process()
		return true
	}

	if atomic.LoadInt64(&u.TimeUsed) >= MAX_TIME {
		return false
	}

	done := make(chan struct{})
	ticker := time.Tick(1 * time.Second)

	// This demonstrates how you can stop a blocking process.
	// process() blocks but if it runs in a goroutine we
	// get past it and then we can return from this function early
	// if we want, thereby skipping the blocking line of code.
	go func() {
		process()
		done <- struct{}{}
	}()

	for {
		select {
		case <-ticker:
			atomic.AddInt64(&u.TimeUsed, 1)
			if u.TimeUsed > MAX_TIME {
				return false
			}
			log.Printf("UserID %d has used %d seconds", u.ID, u.TimeUsed)
		case <-done:
			log.Printf("UserID %d has used %d seconds and has finished streaming a video", u.ID, u.TimeUsed)
			return true
		}
	}
}

func main() {
	RunMockServer()
}

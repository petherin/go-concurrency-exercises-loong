//////////////////////////////////////////////////////////////////////
//
// Given is a SessionManager that stores session information in
// memory. The SessionManager itself is working, however, since we
// keep on adding new sessions to the manager our program will
// eventually run out of memory.
//
// Your task is to implement a session cleaner routine that runs
// concurrently in the background and cleans every session that
// hasn't been updated for more than 5 seconds (of course usually
// session times are much longer).
//
// Note that we expect the session to be removed anytime between 5 and
// 7 seconds after the last update. Also, note that you have to be
// very careful in order to prevent race conditions.
//

package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sessions map[string]Session
	mu       sync.Mutex
}

func (sm *SessionManager) startCleaning() {
	tick := time.Tick(time.Second)
	for {
		<-tick
		sm.mu.Lock()
		for sessionID, session := range sm.sessions {
			if time.Since(session.LastUpdated) > 5*time.Second {
				fmt.Println("deleting session ", sessionID)
				delete(sm.sessions, sessionID)
			}
		}
		sm.mu.Unlock()
	}
}

// Session stores the session's data
type Session struct {
	Data        map[string]interface{}
	LastUpdated time.Time
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions: make(map[string]Session),
	}

	go m.startCleaning()

	return m
}

// CreateSession creates a new session and returns the sessionID
func (m *SessionManager) CreateSession() (string, error) {
	sessionID, err := MakeSessionID()
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[sessionID] = Session{
		Data:        make(map[string]interface{}),
		LastUpdated: time.Now(),
	}

	return sessionID, nil
}

// ErrSessionNotFound returned when sessionID not listed in
// SessionManager
var ErrSessionNotFound = errors.New("SessionID does not exists")

// GetSessionData returns data related to session if sessionID is
// found, errors otherwise
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session.Data, nil
}

// UpdateSessionData overwrites the old session data with the new one
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	_, ok := m.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Hint: you should renew expiry of the session here
	m.sessions[sessionID] = Session{
		Data:        data,
		LastUpdated: time.Now(),
	}

	return nil
}

func processSession(m *SessionManager, key string, value string) string {
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data[key] = value

	err = m.UpdateSessionData(sID, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set", key, "to", value)

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get session data:", updatedData)

	return sID
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sIDs := []string{}

	for i := 0; i < 50; i++ {
		sID := processSession(m, fmt.Sprintf("website%d", i), fmt.Sprintf("longhoang.de%d", i))
		sIDs = append(sIDs, sID)
	}

	updateTicker := time.Tick(100 * time.Millisecond)
	done := make(chan struct{})
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	go func() {
		for {
			select {
			case <-updateTicker:
				if len(sIDs) < 1 {
					// no more sessions ids, exit
					done <- struct{}{}
					return
				}

				randsID := r1.Intn(len(sIDs))
				existing, err := m.GetSessionData(sIDs[randsID])
				if err != nil {
					sIDs = append(sIDs[:randsID], sIDs[randsID+1:]...)
					continue
				}

				for k, v := range existing {
					val := v.(string) + "a"
					existing[k] = val
					log.Println("Update session data, set", k, "to", val)
				}

				err = m.UpdateSessionData(sIDs[randsID], existing)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()

	<-done
}

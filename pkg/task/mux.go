package task

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/hibiken/asynq"
)

/*
* The code for this Mux router is based off the mux router from Asynq
* It has been adapted to better fit Archer's use case
 */

// Handler returns the handler for a given pattern
func (th *TaskHandler) Handler(t *asynq.Task) (h TaskModule, pattern string, err error) {
	// Lock mutex
	th.muxMutex.RLock()
	defer th.muxMutex.RUnlock()

	// Find matching handler
	h, pattern = th.match(t.Type())
	if h == nil {
		return nil, "", errors.New("no matching handler found")
	}

	return
}

// Find a handler on a handler map given a typename string.
// Most-specific (longest) pattern wins.
func (th *TaskHandler) match(typename string) (h TaskModule, pattern string) {
	// Check for exact match first.
	v, ok := th.muxEntries[typename]
	if ok {
		return v.h, v.pattern
	}

	// Check for longest valid match.
	// th.es contains all patterns from longest to shortest.
	for _, e := range th.muxEntriesSorted {
		if strings.HasPrefix(typename, e.pattern) {
			return e.h, e.pattern
		}
	}
	return nil, ""

}

// Handle registers a handler for the given pattern.
func (th *TaskHandler) Handle(pattern string, handler TaskModule) error {
	// Lock mutex
	th.muxMutex.Lock()
	defer th.muxMutex.Unlock()

	// Parameter validation
	if strings.TrimSpace(pattern) == "" {
		return errors.New("pattern cannot be empty")
	}
	if handler == nil {
		return errors.New("handler cannot be null")
	}

	// Check if pattern exists
	if _, exist := th.muxEntries[pattern]; exist {
		return fmt.Errorf("pattern '%s' already exists", pattern)
	}

	// Add mux entry
	e := muxEntry{h: handler, pattern: pattern}
	th.muxEntries[pattern] = e

	// Sort entries
	th.muxEntriesSorted = appendSorted(th.muxEntriesSorted, e)

	return nil
}

// appendSorted sorts mux entries from longest to shortest
func appendSorted(es []muxEntry, e muxEntry) []muxEntry {
	n := len(es)
	i := sort.Search(n, func(i int) bool {
		return len(es[i].pattern) < len(e.pattern)
	})
	if i == n {
		return append(es, e)
	}
	// we now know that i points at where we want to insert.
	es = append(es, muxEntry{}) // try to grow the slice in place, any entry works.
	copy(es[i+1:], es[i:])      // shift shorter entries down.
	es[i] = e
	return es
}

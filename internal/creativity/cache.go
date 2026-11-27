package creativity

import "slices"

type cacheEntry struct {
	request
	last, first, hits int
}

func runCache(trace []request, capacity int, p Policy, mode string) trial {
	if mode == "2q" {
		return runTwoQueue(trace, capacity)
	}
	var out trial
	var resident, history []cacheEntry
	for now, r := range trace {
		idx := slices.IndexFunc(resident, func(e cacheEntry) bool { return e.key == r.key })
		if idx >= 0 {
			out.actions = append(out.actions, 1)
			out.score += r.cost * r.value
			resident[idx].hits++
			resident[idx].last = now
			resident[idx].request = r
			continue
		}
		out.actions = append(out.actions, 0)
		hits := 1
		ghost := slices.IndexFunc(history, func(e cacheEntry) bool { return e.key == r.key })
		if ghost >= 0 {
			hits += history[ghost].hits
			history = slices.Delete(history, ghost, ghost+1)
		}
		// History retains bounded metadata for non-resident misses. An item
		// admitted from history keeps its observed count in its resident slot.
		if mode == "policy" && hits < max(1, p.Admit) {
			if p.History > 0 {
				history = append(history, cacheEntry{request: r, hits: hits})
				if len(history) > p.History*capacity {
					history = history[1:]
				}
				out.history = max(out.history, len(history))
			}
			continue
		}
		if len(resident) == capacity {
			victim := 0
			priority := func(e cacheEntry) int {
				if mode == "lru" {
					return e.last
				}
				return p.rank(8-min(8, now-e.last), min(8, e.hits), e.value, e.cost, min(8, now-e.first))
			}
			worst := priority(resident[0])
			for i, e := range resident[1:] {
				rank := priority(e)
				if rank < worst || (rank == worst && e.last < resident[victim].last) {
					victim, worst = i+1, rank
				}
			}
			resident = slices.Delete(resident, victim, victim+1)
		}
		resident = append(resident, cacheEntry{request: r, last: now, first: now, hits: hits})
		out.resident = max(out.resident, len(resident))
	}
	return out
}

// Full 2Q: A1in is FIFO, Am is LRU, and a hit in the non-resident
// A1out promotes a passage to Am. A1in hits do not promote to Am.
func runTwoQueue(trace []request, capacity int) trial {
	var out trial
	var in, main, ghost []int
	target := max(1, capacity/4)
	for _, r := range trace {
		if pos := slices.Index(main, r.key); pos >= 0 {
			main = slices.Delete(main, pos, pos+1)
			main = append(main, r.key)
			out.actions = append(out.actions, 1)
			out.score += r.cost * r.value
			continue
		}
		if slices.Contains(in, r.key) {
			out.actions = append(out.actions, 1)
			out.score += r.cost * r.value
			continue
		}
		out.actions = append(out.actions, 0)
		pos := slices.Index(ghost, r.key)
		if pos >= 0 {
			ghost = slices.Delete(ghost, pos, pos+1)
		}
		if len(in)+len(main) == capacity {
			if len(in) > target || len(main) == 0 {
				ghost = append(ghost, in[0])
				in = in[1:]
				if len(ghost) > max(1, capacity/2) {
					ghost = ghost[1:]
				}
			} else {
				main = main[1:]
			}
		}
		if pos >= 0 {
			main = append(main, r.key)
		} else {
			in = append(in, r.key)
		}
		out.resident = max(out.resident, len(in)+len(main))
		out.history = max(out.history, len(ghost))
	}
	return out
}

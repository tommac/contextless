package creativity

import "slices"

// Each epoch has a fresh fixed credit allowance. Expired requests earn no
// value. FIFO skips pending reads that cannot fit the remaining allowance;
// value-cost greedily selects the highest value/credit, oldest first on ties.
func runSchedule(epochs [][]job, bound int, p Policy, mode string) trial {
	var out trial
	var pending []job
	var recent []int
	for epoch := 0; epoch < len(epochs) || len(pending) > 0; epoch++ {
		pending = slices.DeleteFunc(pending, func(j job) bool { return epoch > j.expiry })
		if epoch < len(epochs) {
			pending = append(pending, epochs[epoch]...)
			for _, j := range epochs[epoch] {
				recent = append(recent, j.key)
			}
			if len(recent) > 3*bound {
				recent = recent[len(recent)-3*bound:]
			}
			out.history = max(out.history, len(recent))
		}
		rank := func(j job) int {
			recency, freq := 0, 0
			for i, key := range recent {
				if key == j.key {
					freq++
					recency = max(0, 8-(len(recent)-1-i))
				}
			}
			return p.rank(recency, min(8, freq), j.value, j.cost, min(8, epoch-j.arrival))
		}
		credits := bound
		for credits > 0 {
			picked, best := -1, 0
			for i, j := range pending {
				if j.cost > credits {
					continue
				}
				switch mode {
				case "fifo":
					picked = i
				case "value-cost":
					if picked < 0 || j.value*pending[picked].cost > pending[picked].value*j.cost {
						picked = i
					}
				default:
					score := rank(j)
					if picked < 0 || score > best {
						picked, best = i, score
					}
				}
				if mode == "fifo" {
					break
				}
			}
			if picked < 0 {
				break
			}
			j := pending[picked]
			pending = slices.Delete(pending, picked, picked+1)
			credits -= j.cost
			out.score += j.value
			out.actions = append(out.actions, j.id)
		}
		out.actions = append(out.actions, -1)
		out.credits = max(out.credits, bound-credits)
	}
	return out
}

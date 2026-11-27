package creativity

type request struct{ key, value, cost int }
type job struct {
	request
	id, arrival, expiry int
}
type workload struct {
	requests []request
	epochs   [][]job
}

// xorshift32 specifies the generator independently of the runtime's math/rand.
type generator uint32

func (g *generator) next(n int) int {
	x := uint32(*g)
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	*g = generator(x)
	return int(uint64(x) * uint64(n) >> 32)
}

func makeWorkload(scenario int, seed uint32, shape string) workload {
	g := generator(seed + 0x9e3779b9 + uint32(scenario)*0xa511e9b3)
	if g == 0 {
		g = 1
	}
	bound := budgets[scenario]
	hot := max(2, bound-2)
	w := workload{epochs: make([][]job, 36)}
	var recent []int
	id := 0
	// Each scenario changes the distribution, not the evaluator or its rules.
	for epoch := range w.epochs {
		group := 0
		if scenario == 3 {
			group = epoch / 9 * 100
		}
		n := bound + g.next(3)
		if scenario == 2 || scenario == 5 || shape == "burst" {
			if epoch%5 < 2 {
				n = 2*bound + g.next(bound)
			} else {
				n = 2 + g.next(3)
			}
		}
		for i := 0; i < n; i++ {
			kind := g.next(100)
			scanShare := 20
			if scenario == 1 || shape == "scan" {
				scanShare = 55
			}
			if scenario == 2 || scenario == 5 || shape == "burst" {
				scanShare = 8
			}
			var key int
			switch {
			case (scenario == 7 || shape == "reuse") && i%3 == 0 && len(recent) > 2*bound:
				key = recent[len(recent)-2*bound-1]
			case kind < scanShare:
				key = 10000 + id
			case (scenario == 2 || scenario == 5 || shape == "burst") && kind < 75:
				key = group + epoch/4%3*hot + g.next(2)
			default:
				key = group + g.next(hot)
			}
			// Public features known at arrival. Values and costs change the
			// objective independently of whether a key is recent.
			value := 1 + key%3
			cost := 1 + key%2
			if scenario == 4 && key%4 == 0 {
				cost = 3
			}
			if scenario == 6 && id%9 == 0 {
				value = 9
			}
			if scenario == 5 && epoch%5 < 2 {
				value += 2
			}
			life := 1 + g.next(3)
			if scenario == 2 || scenario == 6 {
				life = g.next(3)
			}
			r := request{key: key, value: value, cost: cost}
			w.requests = append(w.requests, r)
			w.epochs[epoch] = append(w.epochs[epoch], job{request: r, id: id, arrival: epoch, expiry: epoch + life})
			recent = append(recent, key)
			id++
		}
	}
	return w
}

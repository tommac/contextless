package providers

import (
	"fmt"
	"strings"
)

// The smoke measures tokenizer constants on fixed fixtures (04 §4.5). Both
// fixtures are generated deterministically from a small vocabulary with a
// linear congruential generator, so they are identical on every machine
// and need no data file.

// ProseFixtureWords is the size of the prose fixture.
const ProseFixtureWords = 5000

var fixtureVocabulary = strings.Fields(`the a an and but or of to in on at by with from for over under
	between through after before while river fence clerk ledger lamp window door harbor road
	morning evening winter spring shadow letter coin bag chain hook key map bread salt wire
	she he they we it her his their our it's old new long short cold warm quiet loud slow quick
	walked turned waited counted measured carried opened closed watched listened answered asked
	said whispered wrote read found lost kept gave took held left stayed crossed followed
	there here again still almost never always once twice today tomorrow yesterday later
	because although until unless whether because since though then now soon`)

type lcg struct{ state uint64 }

func (g *lcg) next() uint64 {
	g.state = g.state*6364136223846793005 + 1442695040888963407
	return g.state >> 33
}

func (g *lcg) pick(n int) int { return int(g.next() % uint64(n)) }

// FillerProse returns n words of deterministic pseudo-prose in sentences
// and paragraphs, for the compaction probe and the prose fixture.
func FillerProse(n int) string {
	g := &lcg{state: 0x5eed}
	var b strings.Builder
	words := 0
	sentence := 0
	for words < n {
		length := 6 + g.pick(9)
		for i := 0; i < length && words < n; i++ {
			w := fixtureVocabulary[g.pick(len(fixtureVocabulary))]
			if i == 0 {
				w = strings.ToUpper(w[:1]) + w[1:]
			}
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(w)
			words++
		}
		b.WriteByte('.')
		sentence++
		if sentence%(4+g.pick(4)) == 0 {
			b.WriteString("\n\n")
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.TrimSpace(b.String()) + "\n"
}

// ProseFixture is the fixed 5,000-word prose fixture.
func ProseFixture() string { return FillerProse(ProseFixtureWords) }

// YAMLFixture is the fixed structured-text fixture: about 24 KB of
// frontmatter-like YAML records.
func YAMLFixture() string {
	g := &lcg{state: 0x7a3f}
	var b strings.Builder
	for i := 0; i < 120; i++ {
		fmt.Fprintf(&b, "- id: ev-%08x\n  type: event\n  title: %s %s %s\n", g.next()&0xffffffff,
			fixtureVocabulary[g.pick(len(fixtureVocabulary))], fixtureVocabulary[g.pick(len(fixtureVocabulary))],
			fixtureVocabulary[g.pick(len(fixtureVocabulary))])
		fmt.Fprintf(&b, "  story_time: {start: \"%d-%02d-%02dT%02d:00\", precision: hour}\n",
			1990+g.pick(30), 1+g.pick(12), 1+g.pick(28), g.pick(24))
		fmt.Fprintf(&b, "  participants: [\"[[ent-%s]]\", \"[[ent-%s]]\"]\n",
			fixtureVocabulary[g.pick(len(fixtureVocabulary))], fixtureVocabulary[g.pick(len(fixtureVocabulary))])
		fmt.Fprintf(&b, "  consequences:\n    - %s\n    - %s\n",
			fillerProseLine(g, 8), fillerProseLine(g, 7))
		fmt.Fprintf(&b, "  told_in: [{scene: sc-%08x, frame: %d, mode: shown}]\n", g.next()&0xffffffff, g.pick(3))
	}
	return b.String()
}

// fillerProseLine returns one short sentence from a generator.
func fillerProseLine(g *lcg, n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = fixtureVocabulary[g.pick(len(fixtureVocabulary))]
	}
	return strings.Join(parts, " ")
}

---
type: result
status: measured
area: packets
needs: none
measured-by: us
---
An engine test builds a conversation of 1,200 turns, about 1.3M tokens and four times the largest window, in which every turn is either in the packet verbatim or named in it. The roadmap describes the scale test, there with several hundred turns, as checking that every packet fits its cap, that the recent turns and the new message are always verbatim, that omitted turns are listed by id, and that a recalled old turn comes back byte for byte. It runs with scripted providers and no model. So it shows that nothing drops out of view when history outgrows every window, not that a model answers well from such a packet.

## Evidence
- `internal/engine/scale_test.go`: "// turns are verbatim, every other turn is either verbatim or named (catalog,"
- `docs/ROADMAP.md`: "checks that every packet fits its cap, that the recent turns and the focus are always verbatim, that omitted turns are listed by id, and that a recalled old turn comes back byte for byte"
- `internal/providers/doc.go`: "// Adapter per CLI (claude, codex, gemini, kimi, grok, muse) plus a scripted"

## Relations
- supports:: [[A record goes in whole or is listed by id]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Each call gets a fresh packet built by code]]
- supports:: [[Packets are sized before the model starts]]
- supports:: [[Catalog cards are an index not a replacement]]
- supports:: [[The plan cap is 60 to 70 percent of the window]]
- supports:: [[No model has yet answered from a packet]]

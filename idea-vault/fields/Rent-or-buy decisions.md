---
type: field
status: external
area: caching
needs: none
---
Rent-or-buy is the online decision, studied in algorithms, of whether to pay a one-time cost now or keep paying a smaller cost each time, without knowing how many times remain. The report frames cache writes this way: writing a prefix costs a premium now and pays back only if later calls arrive inside the cache lifetime. The decision is therefore driven by the gap between messages.

## Evidence
- `reports/Cutting edge context memory research.md`: "which makes it an online rent-or-buy decision driven by the gap between messages"
- `reports/Cutting edge context memory research.md`: "writing a prefix now pays only if later calls arrive inside the lifetime"
- `reports/Cutting edge context memory research.md`: "A write at 1.25x with reads at 0.1x pays back after one read"

## Relations
- analog-of:: [[The priced knapsack with a prefix constraint]]
- analog-of:: [[Message gaps set the cache hit rate]]
- analog-of:: [[Prefix-priced context selection]]
- analog-of:: [[Catalog spine and epoch window make the prefix repeat]]

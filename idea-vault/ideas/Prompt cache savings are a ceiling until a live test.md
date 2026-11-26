---
type: idea
status: designed
area: caching
needs: model-calls
---
The replay computes what the bill would be if every call were served its shared prefix at a tenth of the price, and no call has yet been sent to check. As documented in November 2026, Anthropic and OpenAI's GPT-5.6 and later reuse prefixes only up to block or message boundaries, the brief is one text block, and Claude Code shares a cache across fresh runs only when its startup context matches. So the layout steps stay off until a live test sends A+B then A+C, with A inside one block and as its own block, from one stable directory, and reads the cache fields.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "Those savings are a ceiling. As documented in November 2026, Anthropic and OpenAI's GPT-5.6 and later reuse prefixes only up to block or message boundaries, and the brief is one text block."
- `docs/CONTEXT-COMPILER.md`: "send A+B then A+C, with A as part of one block and as its own block, from one stable directory, and read the cache fields in `usage.json`"

## Relations
- builds-on:: [[The brief runs from stable to volatile]]
- builds-on:: [[Catalog spine and epoch window make the prefix repeat]]
- builds-on:: [[Provider CLIs run as stateless workers]]
- tested-by:: [[Cached at a tenth the replayed input falls from 14381 to 10267]]
- tested-by:: [[Anthropic caches a prefix only through a marked block]]
- tested-by:: [[OpenAI GPT-5.6 reuses a prefix only at message ends or explicit breakpoints]]
- tested-by:: [[Claude Code shares a cache across fresh runs only when startup context matches]]

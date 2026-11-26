// Package mcp exposes ContextlessAI as a first-class Model Context Protocol
// (MCP) stdio JSON-RPC 2.0 server so external agents and IDEs (Claude Code,
// Codex, Cursor, Antigravity) can mount a ContextlessAI vault as their
// lossless, verbatim, snapshot-pinned external memory and context compiler.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/tommac/contextless/internal/engine"
	"github.com/tommac/contextless/internal/packet"
	"github.com/tommac/contextless/internal/store"
	"github.com/tommac/contextless/internal/vault"
)

// Server serves one ContextlessAI Engine over MCP JSON-RPC 2.0 stdio.
type Server struct {
	E *engine.Engine
}

// New creates an MCP server backed by e.
func New(e *engine.Engine) *Server { return &Server{E: e} }

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// Serve processes newline-delimited JSON-RPC 2.0 MCP requests from r and
// writes responses to w until EOF or ctx cancellation.
func (s *Server) Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64<<10), 16<<20)
	enc := json.NewEncoder(w)
	for sc.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			_ = enc.Encode(rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: -32700, Message: "parse error: " + err.Error()}})
			continue
		}
		// JSON-RPC notifications have no id and expect no reply.
		if len(req.ID) == 0 || string(req.ID) == "null" {
			continue
		}
		resp := s.handle(ctx, req)
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
	return sc.Err()
}

func (s *Server) handle(ctx context.Context, req rpcRequest) rpcResponse {
	ok := func(res any) rpcResponse { return rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: res} }
	fail := func(code int, msg string) rpcResponse {
		return rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: code, Message: msg}}
	}
	switch req.Method {
	case "initialize":
		return ok(map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "contextless-ai",
				"version": "0.2.0",
			},
		})
	case "ping":
		return ok(map[string]any{})
	case "tools/list":
		return ok(map[string]any{"tools": toolCatalog()})
	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return fail(-32602, "invalid params: "+err.Error())
		}
		res, isErr := s.callTool(ctx, p.Name, p.Arguments)
		b, _ := json.MarshalIndent(res, "", "  ")
		return ok(map[string]any{
			"content": []map[string]any{{"type": "text", "text": string(b)}},
			"isError": isErr,
		})
	case "resources/list":
		return ok(map[string]any{
			"resources": []map[string]any{
				{"uri": "cx://session", "name": "Workspace Session", "mimeType": "text/markdown"},
				{"uri": "cx://decisions", "name": "Decisions in Force", "mimeType": "application/json"},
				{"uri": "cx://receipts", "name": "Living Answer Receipts", "mimeType": "application/json"},
			},
		})
	case "resources/read":
		var p struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return fail(-32602, "invalid params: "+err.Error())
		}
		text, mime, err := s.readResource(p.URI)
		if err != nil {
			return fail(-32002, err.Error())
		}
		return ok(map[string]any{
			"contents": []map[string]any{{"uri": p.URI, "mimeType": mime, "text": text}},
		})
	default:
		return fail(-32601, "method not found: "+req.Method)
	}
}

func (s *Server) readResource(uri string) (string, string, error) {
	switch {
	case uri == "cx://session":
		sess, err := s.E.V.Session()
		if err != nil {
			return "", "", err
		}
		b, _ := json.MarshalIndent(sess, "", "  ")
		return string(b), "application/json", nil
	case uri == "cx://decisions":
		b, _ := json.MarshalIndent(s.E.V.Decisions(), "", "  ")
		return string(b), "application/json", nil
	case uri == "cx://receipts":
		rs, err := s.E.S.ListReceipts("")
		if err != nil {
			return "", "", err
		}
		b, _ := json.MarshalIndent(rs, "", "  ")
		return string(b), "application/json", nil
	case strings.HasPrefix(uri, "cx://turn/"):
		id := strings.TrimPrefix(uri, "cx://turn/")
		t, err := s.E.V.Turn(id)
		if err != nil {
			return "", "", err
		}
		return t.Body, "text/markdown", nil
	}
	return "", "", fmt.Errorf("unknown resource %q", uri)
}

func (s *Server) callTool(ctx context.Context, name string, raw json.RawMessage) (any, bool) {
	switch name {
	case "cx_compile_packet":
		var a struct {
			Query        string `json:"query"`
			AdaptivePack bool   `json:"adaptive_pack"`
		}
		if err := json.Unmarshal(raw, &a); err != nil || strings.TrimSpace(a.Query) == "" {
			return map[string]string{"error": "query is required"}, true
		}
		oldAdaptive := s.E.Cfg.AdaptivePack
		if a.AdaptivePack {
			s.E.Cfg.AdaptivePack = true
			defer func() { s.E.Cfg.AdaptivePack = oldAdaptive }()
		}
		sized, err := s.E.DryPacket(a.Query)
		if err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		return map[string]any{
			"profile":        sized.Profile.ID,
			"whole_history":  sized.WholeHistory,
			"budget":         sized.Budget,
			"records":        len(sized.Items),
			"omitted":        len(sized.Dropped),
			"omitted_handle": sized.OmittedHandle,
			"index":          sized.Index,
			"brief":          sized.Brief,
		}, false

	case "cx_explain":
		var a struct {
			Query   string   `json:"query"`
			DropIDs []string `json:"drop_ids"`
		}
		if err := json.Unmarshal(raw, &a); err != nil || strings.TrimSpace(a.Query) == "" {
			return map[string]string{"error": "query is required"}, true
		}
		plan, err := s.E.Explain(a.Query, a.DropIDs)
		if err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		return plan, false

	case "cx_recall":
		var a struct {
			IDs     []string          `json:"ids"`
			Queries []string          `json:"queries"`
			Ops     []packet.RecallOp `json:"ops"`
			Op      string            `json:"op"`
			Target  string            `json:"target"`
			Radius  int               `json:"radius"`
			Limit   int               `json:"limit"`
		}
		if err := json.Unmarshal(raw, &a); err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		if a.Op != "" && a.Target != "" {
			a.Ops = append(a.Ops, packet.RecallOp{Op: a.Op, Target: a.Target, Radius: a.Radius, Limit: a.Limit})
		}
		lastSeq, _ := s.E.V.LastSeq()
		focus := vault.Turn{ID: vault.TurnID(lastSeq + 1), Seq: lastSeq + 1}
		opIDs, extras, notes := s.E.ExecuteRecallOps(a.Ops, focus, nil)
		allIDs := append(append([]string{}, a.IDs...), opIDs...)
		if len(a.Queries) > 0 {
			turns, err := s.E.V.Turns()
			if err != nil {
				return map[string]string{"error": err.Error()}, true
			}
			for _, q := range a.Queries {
				for _, r := range packet.RankAtoms(turns, s.E.S, q, focus.Seq, 6) {
					allIDs = append(allIDs, r.Atom.ID())
				}
			}
		}
		type passage struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		}
		var passages []passage
		seen := map[string]bool{}
		for _, id := range allIDs {
			if seen[id] {
				continue
			}
			seen[id] = true
			ref, ok := vault.ParseRef(id)
			if !ok {
				continue
			}
			t, err := s.E.V.Turn(ref.Turn)
			if err != nil {
				continue
			}
			txt := t.Body
			if ref.Atom > 0 {
				if sub, ok := vault.AtomsText(t.ID, t.Body, ref.Atom, ref.AtomEnd); ok {
					txt = sub
				}
			} else if ref.Part > 0 {
				if parts := vault.Parts(t.Body); ref.Part <= len(parts) {
					txt = parts[ref.Part-1]
				}
			}
			passages = append(passages, passage{ID: id, Text: txt})
		}
		return map[string]any{"passages": passages, "extras": extras, "notes": notes}, false

	case "cx_append_turn":
		var a struct {
			Role   string `json:"role"`
			Author string `json:"author"`
			Body   string `json:"body"`
		}
		if err := json.Unmarshal(raw, &a); err != nil || strings.TrimSpace(a.Body) == "" {
			return map[string]string{"error": "body is required"}, true
		}
		role := vault.RoleUser
		if strings.EqualFold(a.Role, "assistant") {
			role = vault.RoleAssistant
		}
		t, err := s.E.V.AppendTurn(role, a.Author, "", "", a.Body)
		if err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		_ = s.E.Reindex()
		alerts, _ := s.E.CheckStaleReceipts(t, nil, nil)
		s.E.V.Commit("MCP append " + t.ID)
		return map[string]any{"turn": t, "stale_alerts": alerts}, false

	case "cx_create_receipt":
		var a struct {
			Question   string                   `json:"question"`
			AnswerText string                   `json:"answer_text"`
			Answer     string                   `json:"answer"`
			AsOfSeq    int                      `json:"as_of_seq"`
			Evidence   []store.ReceiptEvidence  `json:"evidence"`
			Predicates []store.ReceiptPredicate `json:"predicates"`
			Unresolved string                   `json:"unresolved_scope"`
		}
		if err := json.Unmarshal(raw, &a); err != nil || strings.TrimSpace(a.Question) == "" {
			return map[string]string{"error": "question is required"}, true
		}
		if a.AnswerText == "" {
			a.AnswerText = a.Answer
		}
		rcpt, err := s.E.CreateReceipt(a.Question, "", a.AnswerText, a.AsOfSeq, a.Evidence, a.Predicates, a.Unresolved, "")
		if err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		return rcpt, false

	case "cx_watch_status":
		var a struct {
			Status string `json:"status"`
		}
		_ = json.Unmarshal(raw, &a)
		rs, err := s.E.S.ListReceipts(a.Status)
		if err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		return map[string]any{"receipts": rs}, false

	case "cx_diff_since":
		var a struct {
			ReceiptID string `json:"receipt_id"`
		}
		if err := json.Unmarshal(raw, &a); err != nil || strings.TrimSpace(a.ReceiptID) == "" {
			return map[string]string{"error": "receipt_id is required"}, true
		}
		diff, err := s.E.DiffReceiptSince(a.ReceiptID)
		if err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		return diff, false

	case "cx_ledger_query":
		var a struct {
			ResumeID string                 `json:"resume_id"`
			Spec     engine.LedgerQuerySpec `json:"spec"`
			Kind     string                 `json:"kind"`
			Op       string                 `json:"op"`
			Filter   string                 `json:"filter"`
			Status   string                 `json:"status"`
			AsOfSeq  int                    `json:"as_of_seq"`
			Strict   *bool                  `json:"strict"`
		}
		if err := json.Unmarshal(raw, &a); err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		if strings.TrimSpace(a.ResumeID) != "" {
			res, err := s.E.ResumeLedgerQuery(strings.TrimSpace(a.ResumeID))
			if err != nil {
				return map[string]string{"error": err.Error()}, true
			}
			return res, false
		}
		spec := a.Spec
		if spec.Kind == "" {
			spec.Kind = a.Kind
		}
		if spec.Op == "" {
			spec.Op = a.Op
		}
		if spec.Filter == "" {
			spec.Filter = a.Filter
		}
		if spec.Status == "" {
			spec.Status = a.Status
		}
		if spec.AsOfSeq == 0 {
			spec.AsOfSeq = a.AsOfSeq
		}
		if a.Strict != nil {
			spec.Strict = *a.Strict
		}
		res, err := s.E.RunLedgerQuery(spec)
		if err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		return res, false

	case "cx_pack_status":
		var a struct {
			Build      bool                   `json:"build"`
			Verify     bool                   `json:"verify"`
			BlockBytes int                    `json:"block_bytes"`
			Anchors    []vault.EvidenceAnchor `json:"anchors"`
		}
		_ = json.Unmarshal(raw, &a)
		packPath := filepath.Join(s.E.V.Dir, ".cx", "bodies.cxp")
		out := map[string]any{"path": packPath}
		if a.Build {
			st, err := vault.BuildBodyPack(s.E.V, packPath, a.BlockBytes)
			if err != nil {
				return map[string]string{"error": err.Error()}, true
			}
			out["stats"] = st
		}
		if a.Verify || len(a.Anchors) > 0 {
			pk, err := vault.OpenBodyPack(packPath)
			if err != nil {
				return map[string]string{"error": err.Error()}, true
			}
			defer pk.Close()
			if a.Verify {
				if err := pk.VerifyAll(); err != nil {
					return map[string]string{"error": err.Error()}, true
				}
				out["verified"] = true
			}
			if len(a.Anchors) > 0 {
				spans, rst, err := pk.ReadSpans(a.Anchors)
				if err != nil {
					return map[string]string{"error": err.Error()}, true
				}
				out["spans"] = spans
				out["read_stats"] = rst
			}
		}
		return out, false

	case "cx_key_view":
		var a struct {
			Key     string `json:"key"`
			Filter  string `json:"filter"`
			AsOfSeq int    `json:"as_of_seq"`
			Limit   int    `json:"limit"`
		}
		_ = json.Unmarshal(raw, &a)
		if strings.TrimSpace(a.Key) != "" {
			rows, err := s.E.S.QueryKeyHistory(a.Key, a.AsOfSeq)
			if err != nil {
				return map[string]string{"error": err.Error()}, true
			}
			return map[string]any{"key": a.Key, "as_of_seq": a.AsOfSeq, "history": rows}, false
		}
		rows, err := s.E.S.ListActiveKeys(a.Filter, a.AsOfSeq, a.Limit)
		if err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		return map[string]any{"as_of_seq": a.AsOfSeq, "active_keys": rows}, false

	case "cx_fork_workspace":
		var a struct {
			Target  string `json:"target"`
			AsOfSeq int    `json:"as_of_seq"`
			Title   string `json:"title"`
		}
		if err := json.Unmarshal(raw, &a); err != nil || strings.TrimSpace(a.Target) == "" {
			return map[string]string{"error": "target workspace name is required"}, true
		}
		// a target with a path separator is taken as a directory by the
		// configuration, so a tool caller could copy the conversation to
		// any writable place; over MCP a target is a plain workspace name
		if t := strings.TrimSpace(a.Target); !plainName(t) {
			return map[string]string{"error": "target must be a plain workspace name (letters, digits, dot, dash, underscore), not a path"}, true
		}
		fres, err := engine.ForkWorkspace(s.E.Cfg, s.E.Name, strings.TrimSpace(a.Target), engine.ForkOptions{
			AsOfSeq: a.AsOfSeq, Title: a.Title,
		})
		if err != nil {
			return map[string]string{"error": err.Error()}, true
		}
		return fres, false

	case "cx_stats":
		sigs, _ := s.E.S.SignalSummary()
		lastSeq, root, _ := s.E.S.SnapshotRoot(0)
		return map[string]any{
			"workspace":      s.E.Name,
			"snapshot_seq":   lastSeq,
			"snapshot_root":  root,
			"provider_stats": s.E.S.Stats(),
			"signals":        sigs,
		}, false

	default:
		_ = ctx
		return map[string]string{"error": "unknown tool: " + name}, true
	}
}

func toolCatalog() []map[string]any {
	prop := func(typ, desc string) map[string]any { return map[string]any{"type": typ, "description": desc} }
	arrStr := func(desc string) map[string]any {
		return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": desc}
	}
	return []map[string]any{
		{
			"name":        "cx_compile_packet",
			"description": "Compile a deterministic, budget-sized verbatim context packet for a question without compacting or rewriting history.",
			"inputSchema": map[string]any{
				"type": "object", "required": []string{"query"},
				"properties": map[string]any{
					"query":         prop("string", "The question or task to compile context for"),
					"adaptive_pack": prop("boolean", "Use source-diverse anchor-outward excerpt packing"),
				},
			},
		},
		{
			"name":        "cx_explain",
			"description": "Explain the context query plan: query class, lane rankings, token budgets, cache-prefix share, prefix-priced optimizer, and cost estimates.",
			"inputSchema": map[string]any{
				"type": "object", "required": []string{"query"},
				"properties": map[string]any{
					"query":    prop("string", "The question to explain"),
					"drop_ids": arrStr("Optional turn or atom IDs to ablate for causal testing"),
				},
			},
		},
		{
			"name":        "cx_recall",
			"description": "Fetch exact verbatim spans by turn/atom/part IDs, search queries, or typed recall ops (exchange, open_block, timeline, grep).",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"ids":     arrStr("Turn, atom (t000012#a3-5), or part (t000012#p1) IDs"),
					"queries": arrStr("Full-text + value search queries"),
					"ops": map[string]any{
						"type":        "array",
						"description": "Typed recall verbs: exchange, open_block, timeline, grep",
						"items": map[string]any{
							"type": "object", "required": []string{"op", "target"},
							"properties": map[string]any{
								"op":     prop("string", "exchange | open_block | timeline | grep"),
								"target": prop("string", "Turn ID, atom ID, topic, or literal/regex pattern"),
								"radius": prop("integer", "Exchange radius (default 1)"),
								"limit":  prop("integer", "Max grep lines (default 8)"),
							},
						},
					},
				},
			},
		},
		{
			"name":        "cx_append_turn",
			"description": "Append a verbatim turn to the vault, index it, and return any Living Answer staleness alerts triggered by the new record.",
			"inputSchema": map[string]any{
				"type": "object", "required": []string{"body"},
				"properties": map[string]any{
					"role":   prop("string", "user or assistant"),
					"author": prop("string", "Optional author or model id"),
					"body":   prop("string", "Exact verbatim turn text"),
				},
			},
		},
		{
			"name":        "cx_create_receipt",
			"description": "Bind a question and answer to the current (or historical as_of_seq) immutable snapshot with verified evidence spans and watch predicates (including absence predicates).",
			"inputSchema": map[string]any{
				"type": "object", "required": []string{"question"},
				"properties": map[string]any{
					"question":         prop("string", "The question being answered"),
					"answer_text":      prop("string", "The verified answer at this snapshot"),
					"as_of_seq":        prop("integer", "0 for live watch; >0 for historical cutoff"),
					"unresolved_scope": prop("string", "Explicit note of any unchecked scope"),
				},
			},
		},
		{
			"name":        "cx_watch_status",
			"description": "List Living Answer Query Receipts, optionally filtered by status (current, stale, superseded).",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{"status": prop("string", "Optional status filter")},
			},
		},
		{
			"name":        "cx_diff_since",
			"description": "Verify a Query Receipt's original evidence byte-for-byte and return only the new delta turns arriving after its snapshot.",
			"inputSchema": map[string]any{
				"type": "object", "required": []string{"receipt_id"},
				"properties": map[string]any{"receipt_id": prop("string", "Query Receipt ID")},
			},
		},
		{
			"name":        "cx_ledger_query",
			"description": "Run or resume an exact deterministic scoped query over decisions, facts, or turns (count, list, latest, presence, absent) with AsOfSeq time-travel reconstruction and strict withholding when index lags.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"resume_id": prop("string", "Existing ledger query ID to resume incrementally"),
					"kind":      prop("string", "decisions | facts | turns"),
					"op":        prop("string", "count | list | latest | presence | absent"),
					"filter":    prop("string", "Optional topic/keyword filter"),
					"status":    prop("string", "active | superseded | all"),
					"as_of_seq": prop("integer", "0 for live; >0 for historical point-in-time cutoff"),
					"strict":    prop("boolean", "Withhold verdict if unindexed turns exist in scope"),
				},
			},
		},
		{
			"name":        "cx_pack_status",
			"description": "Build, verify, or read coalesced byte-exact spans from the 4 KiB block-compressed lossless body pack (.cx/bodies.cxp).",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"build":       prop("boolean", "Build/refresh .cx/bodies.cxp from turns"),
					"verify":      prop("boolean", "Verify every block and turn SHA-256"),
					"block_bytes": prop("integer", "Block size in bytes (default 4096)"),
				},
			},
		},
		{
			"name":        "cx_key_view",
			"description": "Experimental. Query key=value views over the turn sequence (active keys or the version chain of one key, optionally as of as_of_seq). Known defects are listed in docs/ROADMAP.md: check a value against the turns it cites.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key":       prop("string", "Specific key to inspect full chronological version chain"),
					"filter":    prop("string", "Substring filter when listing active keys"),
					"as_of_seq": prop("integer", "0 for live; >0 for historical point-in-time state"),
					"limit":     prop("integer", "Max active keys to return"),
				},
			},
		},
		{
			"name":        "cx_fork_workspace",
			"description": "Experimental. Fork this workspace, live or as of as_of_seq, into a new isolated git workspace. Known defect: thread and entity records from after the cutoff are copied.",
			"inputSchema": map[string]any{
				"type": "object", "required": []string{"target"},
				"properties": map[string]any{
					"target":    prop("string", "Name of the new forked workspace"),
					"as_of_seq": prop("integer", "0 for full copy; >0 to branch at historical turn seq"),
					"title":     prop("string", "Optional title for forked workspace"),
				},
			},
		},
		{
			"name":        "cx_stats",
			"description": "Return workspace snapshot root, provider run statistics, and free-signal telemetry counts.",
			"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}
}

// plainName reports whether s can only name a workspace under the
// configured root: no separator, no parent reference, no leading dot.
func plainName(s string) bool {
	if s == "" || len(s) > 100 || s[0] == '.' {
		return false
	}
	for _, r := range s {
		ok := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '-' || r == '_'
		if !ok {
			return false
		}
	}
	return !strings.Contains(s, "..")
}

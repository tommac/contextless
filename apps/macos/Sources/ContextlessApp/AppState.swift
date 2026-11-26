import Foundation
import Observation

enum MainTab: String, CaseIterable, Identifiable {
    case chat = "Chat"
    case planner = "Context Planner"
    case vault = "Vault & Views"

    var id: String { rawValue }
    var icon: String {
        switch self {
        case .chat: return "bubble.left.and.bubble.right"
        case .planner: return "cpu"
        case .vault: return "cylinder.split.1x2"
        }
    }
}

@Observable
@MainActor
final class AppState {
    var cxBinaryPath: String = CXRunner.detectBinary()
    var workspaces: [WorkspaceSummary] = []
    var selectedWorkspaceID: String?
    var searchText: String = ""
    var selectedTab: MainTab = .chat

    // Selected workspace content
    var turns: [TurnRecord] = []
    var decisions: [DecisionRecord] = []

    // Chat state
    var draftMessage: String = ""
    var isSending: Bool = false
    var statusMessage: String = ""
    var errorMessage: String?

    // Context Planner (cx explain) state
    var explainQuery: String = ""
    var explainAdaptive: Bool = true
    var explainSessionRoute: Bool = false
    var explainOutput: String = ""
    var isExplaining: Bool = false

    // Vault & Views CLI outputs
    var viewsOutput: String = ""
    var receiptsOutput: String = ""
    var notesOutput: String = ""
    var isLoadingVaultDetails: Bool = false

    // Sheets
    var showNewWorkspaceSheet: Bool = false
    var showForkSheet: Bool = false
    var showDemoSheet: Bool = false
    var demoOutput: String = ""
    var isRunningDemo: Bool = false

    var selectedWorkspace: WorkspaceSummary? {
        guard let id = selectedWorkspaceID else { return nil }
        return workspaces.first(where: { $0.id == id })
    }

    var filteredWorkspaces: [WorkspaceSummary] {
        let q = searchText.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if q.isEmpty { return workspaces }
        return workspaces.filter {
            $0.name.lowercased().contains(q) ||
            $0.title.lowercased().contains(q) ||
            $0.purpose.lowercased().contains(q)
        }
    }

    func refreshWorkspaces(keepSelection: Bool = true) {
        let loaded = VaultReader.listWorkspaces()
        workspaces = loaded
        if !keepSelection || selectedWorkspaceID == nil || !loaded.contains(where: { $0.id == selectedWorkspaceID }) {
            selectedWorkspaceID = loaded.first?.id
        }
        reloadSelectedWorkspace()
    }

    func reloadSelectedWorkspace() {
        guard let ws = selectedWorkspace else {
            turns = []
            decisions = []
            return
        }
        turns = VaultReader.loadTurns(workspacePath: ws.path)
        decisions = VaultReader.loadDecisions(workspacePath: ws.path)
        if selectedTab == .vault {
            Task { await refreshVaultCLIViews() }
        }
    }

    func sendMessage() async {
        guard let ws = selectedWorkspace else { return }
        let msg = draftMessage.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !msg.isEmpty, !isSending else { return }

        isSending = true
        errorMessage = nil
        statusMessage = "Routing through ContextlessAI…"
        draftMessage = ""

        let res = await CXRunner.run(binary: cxBinaryPath, arguments: ["say", ws.name, msg])
        isSending = false
        if res.exitCode == 0 {
            statusMessage = "Saved verbatim to \(ws.name)"
            refreshWorkspaces(keepSelection: true)
        } else {
            draftMessage = msg
            errorMessage = res.combinedOutput.isEmpty ? "cx say exited with status \(res.exitCode)" : res.combinedOutput
            statusMessage = ""
        }
    }

    func explainDraftOrQuery(queryOverride: String? = nil) async {
        guard let ws = selectedWorkspace else { return }
        let q = (queryOverride ?? explainQuery).trimmingCharacters(in: .whitespacesAndNewlines)
        guard !q.isEmpty else { return }

        explainQuery = q
        selectedTab = .planner
        isExplaining = true
        errorMessage = nil

        var args = ["explain", ws.name]
        if explainAdaptive { args.append("-adaptive") }
        if explainSessionRoute { args.append("-session-route") }
        args.append(q)

        let res = await CXRunner.run(binary: cxBinaryPath, arguments: args)
        isExplaining = false
        if res.exitCode == 0 {
            explainOutput = res.stdout
        } else {
            explainOutput = res.combinedOutput
        }
    }

    func refreshVaultCLIViews() async {
        guard let ws = selectedWorkspace else { return }
        isLoadingVaultDetails = true
        async let v = CXRunner.run(binary: cxBinaryPath, arguments: ["views", ws.name])
        async let r = CXRunner.run(binary: cxBinaryPath, arguments: ["stale", ws.name, "-all"])
        async let n = CXRunner.run(binary: cxBinaryPath, arguments: ["notes", ws.name])
        let (vRes, rRes, nRes) = await (v, r, n)
        viewsOutput = vRes.combinedOutput.isEmpty ? "No active materialized key assertions." : vRes.combinedOutput
        receiptsOutput = rRes.combinedOutput.isEmpty ? "No Living Answer receipts recorded yet." : rRes.combinedOutput
        notesOutput = nRes.combinedOutput.isEmpty ? "No Atlas notes imported yet." : nRes.combinedOutput
        isLoadingVaultDetails = false
    }

    func createWorkspace(name: String, title: String, purpose: String) async -> Bool {
        let cleanName = name.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !cleanName.isEmpty else { return false }
        var args = ["new", cleanName]
        if !title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            args += ["-title", title.trimmingCharacters(in: .whitespacesAndNewlines)]
        }
        if !purpose.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            args += ["-purpose", purpose.trimmingCharacters(in: .whitespacesAndNewlines)]
        }
        let res = await CXRunner.run(binary: cxBinaryPath, arguments: args)
        if res.exitCode == 0 {
            refreshWorkspaces(keepSelection: false)
            selectedWorkspaceID = cleanName
            reloadSelectedWorkspace()
            statusMessage = "Created workspace \(cleanName)"
            return true
        } else {
            errorMessage = res.combinedOutput
            return false
        }
    }

    func forkWorkspace(destName: String, asOfSeq: Int, title: String) async -> Bool {
        guard let ws = selectedWorkspace else { return false }
        let cleanDst = destName.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !cleanDst.isEmpty else { return false }
        var args = ["fork", ws.name, cleanDst]
        if asOfSeq > 0 {
            args += ["-as-of", String(asOfSeq)]
        }
        if !title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
            args += ["-title", title.trimmingCharacters(in: .whitespacesAndNewlines)]
        }
        let res = await CXRunner.run(binary: cxBinaryPath, arguments: args)
        if res.exitCode == 0 {
            refreshWorkspaces(keepSelection: false)
            selectedWorkspaceID = cleanDst
            reloadSelectedWorkspace()
            statusMessage = "Forked \(ws.name) → \(cleanDst)"
            return true
        } else {
            errorMessage = res.combinedOutput
            return false
        }
    }

    func runDemo() async {
        showDemoSheet = true
        isRunningDemo = true
        demoOutput = "Running model-free `cx demo`…"
        let res = await CXRunner.run(binary: cxBinaryPath, arguments: ["demo"])
        isRunningDemo = false
        demoOutput = res.combinedOutput
    }
}

import SwiftUI

struct ContentView: View {
    @Bindable var state: AppState

    var body: some View {
        NavigationSplitView {
            SidebarView(state: state)
                .navigationSplitViewColumnWidth(min: 220, ideal: 260, max: 340)
        } detail: {
            if let ws = state.selectedWorkspace {
                WorkspaceDetailView(state: state, workspace: ws)
            } else {
                ContentUnavailableView(
                    "No Workspace Selected",
                    systemImage: "tray",
                    description: Text("Select a workspace on the left or create a new verbatim vault.")
                )
            }
        }
        .toolbar {
            ToolbarItem(placement: .navigation) {
                Button {
                    state.showNewWorkspaceSheet = true
                } label: {
                    Label("New Workspace", systemImage: "plus")
                }
                .help("Create a new ContextlessAI workspace (⌘N)")
            }

            ToolbarItem(placement: .principal) {
                Picker("View", selection: $state.selectedTab) {
                    ForEach(MainTab.allCases) { tab in
                        Label(tab.rawValue, systemImage: tab.icon).tag(tab)
                    }
                }
                .pickerStyle(.segmented)
                .frame(width: 340)
            }

            ToolbarItemGroup(placement: .primaryAction) {
                Button {
                    Task { await state.runDemo() }
                } label: {
                    Label("Run Demo", systemImage: "play.circle")
                }
                .help("Run model-free `cx demo` showcase")

                Button {
                    state.refreshWorkspaces(keepSelection: true)
                } label: {
                    Label("Refresh", systemImage: "arrow.clockwise")
                }
                .help("Reload workspaces from disk (⌘R)")
            }
        }
        .sheet(isPresented: $state.showNewWorkspaceSheet) {
            NewWorkspaceSheet(state: state)
        }
        .sheet(isPresented: $state.showForkSheet) {
            if let ws = state.selectedWorkspace {
                ForkWorkspaceSheet(state: state, source: ws)
            }
        }
        .sheet(isPresented: $state.showDemoSheet) {
            DemoOutputSheet(state: state)
        }
        .onChange(of: state.selectedWorkspaceID) { _, _ in
            state.reloadSelectedWorkspace()
        }
        .onChange(of: state.selectedTab) { _, newTab in
            if newTab == .vault {
                Task { await state.refreshVaultCLIViews() }
            }
        }
        .onAppear {
            state.refreshWorkspaces(keepSelection: false)
        }
    }
}

private struct SidebarView: View {
    @Bindable var state: AppState

    var body: some View {
        List(selection: $state.selectedWorkspaceID) {
            Section("Workspaces (\(state.filteredWorkspaces.count))") {
                ForEach(state.filteredWorkspaces) { ws in
                    HStack {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(ws.title)
                                .font(.headline)
                                .lineLimit(1)
                            Text(ws.purpose.isEmpty ? ws.name : ws.purpose)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                                .lineLimit(1)
                        }
                        Spacer()
                        Text("\(ws.turnCount)")
                            .font(.caption.monospacedDigit())
                            .padding(.horizontal, 6)
                            .padding(.vertical, 2)
                            .background(.quaternary, in: Capsule())
                    }
                    .padding(.vertical, 2)
                    .tag(ws.id)
                }
            }
        }
        .searchable(text: $state.searchText, placement: .sidebar, prompt: "Filter workspaces")
        .safeAreaInset(edge: .bottom) {
            HStack(spacing: 6) {
                Circle()
                    .fill(FileManager.default.isExecutableFile(atPath: state.cxBinaryPath) ? Color.green : Color.orange)
                    .frame(width: 7, height: 7)
                Text(URL(fileURLWithPath: state.cxBinaryPath).lastPathComponent)
                    .font(.caption.monospaced())
                    .foregroundStyle(.secondary)
                Spacer()
                Text("Zero-Compaction Vault")
                    .font(.caption2)
                    .foregroundStyle(.tertiary)
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .background(.bar)
        }
    }
}

private struct WorkspaceDetailView: View {
    @Bindable var state: AppState
    let workspace: WorkspaceSummary

    var body: some View {
        VStack(spacing: 0) {
            if let err = state.errorMessage {
                HStack(spacing: 8) {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .foregroundStyle(.yellow)
                    Text(err)
                        .font(.caption)
                        .lineLimit(3)
                        .textSelection(.enabled)
                    Spacer()
                    Button("Dismiss") { state.errorMessage = nil }
                        .buttonStyle(.borderless)
                        .font(.caption)
                }
                .padding(.horizontal, 14)
                .padding(.vertical, 8)
                .background(Color.red.opacity(0.12))
                Divider()
            }

            switch state.selectedTab {
            case .chat:
                ChatPaneView(state: state, workspace: workspace)
            case .planner:
                PlannerPaneView(state: state, workspace: workspace)
            case .vault:
                VaultPaneView(state: state, workspace: workspace)
            }
        }
        .navigationTitle(workspace.title)
        .navigationSubtitle("\(workspace.name) · \(state.turns.count) verbatim turns · \(state.decisions.filter { $0.status == "active" }.count) active decisions")
    }
}

private struct ChatPaneView: View {
    @Bindable var state: AppState
    let workspace: WorkspaceSummary

    var body: some View {
        VStack(spacing: 0) {
            if state.turns.isEmpty {
                ContentUnavailableView(
                    "Empty Verbatim Vault",
                    systemImage: "text.bubble",
                    description: Text("Send a message below to append turn t000001 and route it through ContextlessAI.")
                )
                .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollViewReader { proxy in
                    ScrollView {
                        LazyVStack(alignment: .leading, spacing: 12) {
                            ForEach(state.turns) { turn in
                                TurnBubbleView(turn: turn)
                                    .id(turn.id)
                            }
                        }
                        .padding(16)
                    }
                    .onChange(of: state.turns.count) { _, _ in
                        if let last = state.turns.last {
                            withAnimation { proxy.scrollTo(last.id, anchor: .bottom) }
                        }
                    }
                    .onAppear {
                        if let last = state.turns.last {
                            proxy.scrollTo(last.id, anchor: .bottom)
                        }
                    }
                }
            }

            Divider()

            VStack(spacing: 8) {
                HStack(alignment: .bottom, spacing: 10) {
                    TextEditor(text: $state.draftMessage)
                        .font(.body)
                        .frame(minHeight: 44, maxHeight: 120)
                        .scrollContentBackground(.hidden)
                        .padding(8)
                        .background(Color(nsColor: .textBackgroundColor), in: RoundedRectangle(cornerRadius: 10))
                        .overlay(
                            RoundedRectangle(cornerRadius: 10)
                                .stroke(Color.secondary.opacity(0.25), lineWidth: 1)
                        )
                        .overlay(alignment: .topLeading) {
                            if state.draftMessage.isEmpty {
                                Text("Message \(workspace.title)… (⌘↩ to send)")
                                    .foregroundStyle(.tertiary)
                                    .padding(.horizontal, 13)
                                    .padding(.vertical, 10)
                                    .allowsHitTesting(false)
                            }
                        }

                    VStack(spacing: 6) {
                        Button {
                            Task { await state.sendMessage() }
                        } label: {
                            if state.isSending {
                                ProgressView().controlSize(.small).frame(width: 72)
                            } else {
                                Label("Send", systemImage: "paperplane.fill")
                                    .frame(width: 72)
                            }
                        }
                        .buttonStyle(.borderedProminent)
                        .keyboardShortcut(.return, modifiers: [.command])
                        .disabled(state.isSending || state.draftMessage.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)

                        Button {
                            let q = state.draftMessage
                            Task { await state.explainDraftOrQuery(queryOverride: q) }
                        } label: {
                            Label("Explain", systemImage: "magnifyingglass")
                                .frame(width: 72)
                        }
                        .buttonStyle(.bordered)
                        .disabled(state.draftMessage.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
                        .help("Dry-run Context Query Planner (`cx explain`) with zero model calls")
                    }
                }

                if !state.statusMessage.isEmpty {
                    HStack {
                        Text(state.statusMessage)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        Spacer()
                    }
                }
            }
            .padding(12)
            .background(.bar)
        }
    }
}

private struct TurnBubbleView: View {
    let turn: TurnRecord
    @State private var showMeta = false

    private var isUser: Bool { turn.role.lowercased() == "user" }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 8) {
                Text(turn.id)
                    .font(.caption.monospaced().bold())
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(isUser ? Color.accentColor.opacity(0.18) : Color.secondary.opacity(0.16), in: RoundedRectangle(cornerRadius: 5))

                Text(isUser ? "User" : (turn.provider.isEmpty ? "Assistant" : turn.provider))
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(isUser ? Color.accentColor : .primary)

                if !turn.created.isEmpty {
                    Text(turn.created)
                        .font(.caption2.monospacedDigit())
                        .foregroundStyle(.tertiary)
                }

                Spacer()

                if !turn.bodySHA.isEmpty {
                    Button {
                        showMeta.toggle()
                    } label: {
                        Image(systemName: "checkmark.seal")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                    .buttonStyle(.plain)
                    .help("SHA-256 verbatim integrity anchor: \(turn.bodySHA)")
                }
            }

            Text(turn.body)
                .font(.body)
                .textSelection(.enabled)
                .frame(maxWidth: .infinity, alignment: .leading)

            if showMeta {
                Text("body_sha: \(turn.bodySHA)")
                    .font(.caption2.monospaced())
                    .foregroundStyle(.secondary)
                    .textSelection(.enabled)
                    .padding(.top, 2)
            }
        }
        .padding(12)
        .background(
            isUser
                ? Color.accentColor.opacity(0.06)
                : Color(nsColor: .controlBackgroundColor),
            in: RoundedRectangle(cornerRadius: 12)
        )
        .overlay(
            RoundedRectangle(cornerRadius: 12)
                .stroke(Color.secondary.opacity(0.15), lineWidth: 1)
        )
    }
}

private struct PlannerPaneView: View {
    @Bindable var state: AppState
    let workspace: WorkspaceSummary

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 10) {
                TextField("Ask a question to inspect packet compilation (`cx explain`)…", text: $state.explainQuery)
                    .textFieldStyle(.roundedBorder)
                    .onSubmit {
                        Task { await state.explainDraftOrQuery() }
                    }

                Toggle("Adaptive", isOn: $state.explainAdaptive)
                    .toggleStyle(.checkbox)

                Toggle("Session Route", isOn: $state.explainSessionRoute)
                    .toggleStyle(.checkbox)

                Button {
                    Task { await state.explainDraftOrQuery() }
                } label: {
                    if state.isExplaining {
                        ProgressView().controlSize(.small)
                    } else {
                        Text("Explain")
                    }
                }
                .buttonStyle(.borderedProminent)
                .disabled(state.isExplaining || state.explainQuery.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
            }

            GroupBox("Context Query Plan (Model-Free)") {
                ScrollView {
                    Text(state.explainOutput.isEmpty
                         ? "Enter a query above and click Explain to inspect query class, lane hits, cache prefix, budget allocation, and optimality bounds without spending model calls."
                         : state.explainOutput)
                        .font(.system(.callout, design: .monospaced))
                        .textSelection(.enabled)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(8)
                }
            }
        }
        .padding(16)
    }
}

private struct VaultPaneView: View {
    @Bindable var state: AppState
    let workspace: WorkspaceSummary

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                HStack {
                    Button {
                        Task { await state.refreshVaultCLIViews() }
                    } label: {
                        Label("Refresh Views", systemImage: "arrow.clockwise")
                    }
                    .disabled(state.isLoadingVaultDetails)

                    Button {
                        state.showForkSheet = true
                    } label: {
                        Label("Fork at Turn…", systemImage: "arrow.triangle.branch")
                    }

                    Spacer()

                    if state.isLoadingVaultDetails {
                        ProgressView().controlSize(.small)
                    }
                }

                GroupBox("Decisions (\(state.decisions.count))") {
                    if state.decisions.isEmpty {
                        Text("No decisions recorded in this workspace yet.")
                            .foregroundStyle(.secondary)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .padding(8)
                    } else {
                        VStack(alignment: .leading, spacing: 8) {
                            ForEach(state.decisions) { d in
                                VStack(alignment: .leading, spacing: 2) {
                                    HStack {
                                        Text(d.status.uppercased())
                                            .font(.caption2.bold())
                                            .padding(.horizontal, 6)
                                            .padding(.vertical, 2)
                                            .background(d.status == "active" ? Color.green.opacity(0.2) : Color.secondary.opacity(0.2), in: Capsule())
                                        Text(d.id)
                                            .font(.caption.monospaced())
                                            .foregroundStyle(.secondary)
                                        if !d.source.isEmpty {
                                            Text("(\(d.source))")
                                                .font(.caption.monospaced())
                                                .foregroundStyle(.tertiary)
                                        }
                                    }
                                    Text(d.statement)
                                        .font(.callout)
                                        .textSelection(.enabled)
                                    if !d.quote.isEmpty {
                                        Text("“\(d.quote)”")
                                            .font(.caption)
                                            .foregroundStyle(.secondary)
                                    }
                                }
                                .padding(.vertical, 4)
                                Divider()
                            }
                        }
                        .padding(6)
                    }
                }

                GroupBox("Materialized Key-Value Views (`cx views`)") {
                    Text(state.viewsOutput)
                        .font(.system(.caption, design: .monospaced))
                        .textSelection(.enabled)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(8)
                }

                GroupBox("Atlas Notes (`cx notes`)") {
                    Text(state.notesOutput)
                        .font(.system(.caption, design: .monospaced))
                        .textSelection(.enabled)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(8)
                }

                GroupBox("Living Answer Receipts (`cx stale -all`)") {
                    Text(state.receiptsOutput)
                        .font(.system(.caption, design: .monospaced))
                        .textSelection(.enabled)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(8)
                }
            }
            .padding(16)
        }
    }
}

private struct NewWorkspaceSheet: View {
    @Bindable var state: AppState
    @Environment(\.dismiss) private var dismiss
    @State private var name = ""
    @State private var title = ""
    @State private var purpose = ""
    @State private var isCreating = false

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("New ContextlessAI Workspace")
                .font(.title2.bold())

            Form {
                TextField("Workspace ID (e.g. kernel-design)", text: $name)
                TextField("Title (optional)", text: $title)
                TextField("Purpose (optional)", text: $purpose)
            }
            .formStyle(.grouped)

            HStack {
                Spacer()
                Button("Cancel") { dismiss() }
                    .keyboardShortcut(.cancelAction)
                Button("Create") {
                    isCreating = true
                    Task {
                        let ok = await state.createWorkspace(name: name, title: title, purpose: purpose)
                        isCreating = false
                        if ok { dismiss() }
                    }
                }
                .buttonStyle(.borderedProminent)
                .keyboardShortcut(.defaultAction)
                .disabled(isCreating || name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
            }
        }
        .padding(20)
        .frame(width: 440)
    }
}

private struct ForkWorkspaceSheet: View {
    @Bindable var state: AppState
    let source: WorkspaceSummary
    @Environment(\.dismiss) private var dismiss
    @State private var destName = ""
    @State private var asOfSeqText = ""
    @State private var title = ""
    @State private var isForking = false

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("Fork Workspace “\(source.name)”")
                .font(.title2.bold())
            Text("Create an independent branch of this vault, optionally rolled back to an earlier turn sequence.")
                .font(.caption)
                .foregroundStyle(.secondary)

            Form {
                TextField("New Workspace ID", text: $destName)
                TextField("As-Of Turn Seq (blank = latest, max \(source.turnCount))", text: $asOfSeqText)
                TextField("Title (optional)", text: $title)
            }
            .formStyle(.grouped)

            HStack {
                Spacer()
                Button("Cancel") { dismiss() }
                    .keyboardShortcut(.cancelAction)
                Button("Fork") {
                    isForking = true
                    let seq = Int(asOfSeqText.trimmingCharacters(in: .whitespacesAndNewlines)) ?? 0
                    Task {
                        let ok = await state.forkWorkspace(destName: destName, asOfSeq: seq, title: title)
                        isForking = false
                        if ok { dismiss() }
                    }
                }
                .buttonStyle(.borderedProminent)
                .keyboardShortcut(.defaultAction)
                .disabled(isForking || destName.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
            }
        }
        .padding(20)
        .frame(width: 460)
    }
}

private struct DemoOutputSheet: View {
    @Bindable var state: AppState
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("ContextlessAI Model-Free Demo (`cx demo`)")
                    .font(.title3.bold())
                Spacer()
                if state.isRunningDemo {
                    ProgressView().controlSize(.small)
                }
                Button("Done") { dismiss() }
                    .keyboardShortcut(.defaultAction)
            }
            Divider()
            ScrollView {
                Text(state.demoOutput)
                    .font(.system(.callout, design: .monospaced))
                    .textSelection(.enabled)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(8)
            }
            .background(Color(nsColor: .textBackgroundColor), in: RoundedRectangle(cornerRadius: 8))
        }
        .padding(18)
        .frame(width: 760, height: 520)
    }
}

import AppKit
import SwiftUI

@main
struct ContextlessApp: App {
    @State private var state = AppState()

    init() {
        NSApplication.shared.setActivationPolicy(.regular)
        NSApplication.shared.activate(ignoringOtherApps: true)
    }

    var body: some Scene {
        WindowGroup("ContextlessAI") {
            ContentView(state: state)
                .frame(minWidth: 920, minHeight: 620)
        }
        .defaultSize(width: 1120, height: 740)
        .commands {
            CommandGroup(replacing: .newItem) {
                Button("New Workspace…") {
                    state.showNewWorkspaceSheet = true
                }
                .keyboardShortcut("n", modifiers: [.command])
            }
            CommandMenu("Vault") {
                Button("Refresh Workspaces") {
                    state.refreshWorkspaces(keepSelection: true)
                }
                .keyboardShortcut("r", modifiers: [.command])

                Button("Run Model-Free Demo") {
                    Task { await state.runDemo() }
                }
                .keyboardShortcut("d", modifiers: [.command, .shift])
            }
        }
    }
}

import Foundation

struct WorkspaceSummary: Identifiable, Hashable, Sendable {
    var id: String { name }
    let name: String
    let title: String
    let purpose: String
    let path: URL
    let turnCount: Int
    let modifiedAt: Date
}

struct TurnRecord: Identifiable, Hashable, Sendable {
    let id: String
    let seq: Int
    let role: String
    let provider: String
    let created: String
    let bodySHA: String
    let body: String
}

struct DecisionRecord: Identifiable, Hashable, Sendable {
    let id: String
    let status: String
    let source: String
    let statement: String
    let quote: String
}

struct CommandResult: Sendable {
    let exitCode: Int32
    let stdout: String
    let stderr: String

    var combinedOutput: String {
        let out = stdout.trimmingCharacters(in: .whitespacesAndNewlines)
        let err = stderr.trimmingCharacters(in: .whitespacesAndNewlines)
        if !out.isEmpty && !err.isEmpty {
            return out + "\n\n" + err
        }
        return out.isEmpty ? err : out
    }
}

enum FrontmatterParser {
    static func parse(_ raw: String) -> (meta: [String: String], body: String) {
        let normalized = raw.replacingOccurrences(of: "\r\n", with: "\n")
        guard normalized.hasPrefix("---\n") else {
            return ([:], normalized.trimmingCharacters(in: .whitespacesAndNewlines))
        }
        let remainder = normalized.dropFirst(4)
        guard let endRange = remainder.range(of: "\n---\n") ?? remainder.range(of: "\n---") else {
            return ([:], normalized.trimmingCharacters(in: .whitespacesAndNewlines))
        }
        let header = String(remainder[..<endRange.lowerBound])
        let bodyStart = endRange.upperBound
        let body = String(remainder[bodyStart...]).trimmingCharacters(in: .whitespacesAndNewlines)

        var meta: [String: String] = [:]
        for line in header.split(separator: "\n", omittingEmptySubsequences: true) {
            guard let colon = line.firstIndex(of: ":") else { continue }
            let key = line[..<colon].trimmingCharacters(in: .whitespaces)
            var val = line[line.index(after: colon)...].trimmingCharacters(in: .whitespaces)
            if (val.hasPrefix("\"") && val.hasSuffix("\"")) || (val.hasPrefix("'") && val.hasSuffix("'")) {
                if val.count >= 2 {
                    val = String(val.dropFirst().dropLast())
                }
            }
            if !key.isEmpty {
                meta[key] = val
            }
        }
        return (meta, body)
    }
}

enum VaultReader {
    static var workspacesRoot: URL {
        FileManager.default.homeDirectoryForCurrentUser
            .appendingPathComponent(".contextless", isDirectory: true)
            .appendingPathComponent("workspaces", isDirectory: true)
    }

    static func listWorkspaces() -> [WorkspaceSummary] {
        let fm = FileManager.default
        guard let dirs = try? fm.contentsOfDirectory(
            at: workspacesRoot,
            includingPropertiesForKeys: [.isDirectoryKey, .contentModificationDateKey],
            options: [.skipsHiddenFiles]
        ) else {
            return []
        }

        var results: [WorkspaceSummary] = []
        for dir in dirs {
            guard (try? dir.resourceValues(forKeys: [.isDirectoryKey]).isDirectory) == true else {
                continue
            }
            let sessionURL = dir.appendingPathComponent("session.md")
            let rawSession = (try? String(contentsOf: sessionURL, encoding: .utf8)) ?? ""
            let parsed = FrontmatterParser.parse(rawSession)
            let name = dir.lastPathComponent
            let title = parsed.meta["title"]?.isEmpty == false ? parsed.meta["title"]! : name
            let purpose = parsed.meta["purpose"] ?? ""

            let turnsDir = dir.appendingPathComponent("turns", isDirectory: true)
            let turnFiles = (try? fm.contentsOfDirectory(at: turnsDir, includingPropertiesForKeys: [.contentModificationDateKey], options: [.skipsHiddenFiles])) ?? []
            let mdTurns = turnFiles.filter { $0.pathExtension == "md" }

            var latestDate = (try? dir.resourceValues(forKeys: [.contentModificationDateKey]).contentModificationDate) ?? .distantPast
            for tf in mdTurns {
                if let d = try? tf.resourceValues(forKeys: [.contentModificationDateKey]).contentModificationDate, d > latestDate {
                    latestDate = d
                }
            }

            results.append(WorkspaceSummary(
                name: name,
                title: title,
                purpose: purpose,
                path: dir,
                turnCount: mdTurns.count,
                modifiedAt: latestDate
            ))
        }

        return results.sorted { $0.modifiedAt > $1.modifiedAt }
    }

    static func loadTurns(workspacePath: URL) -> [TurnRecord] {
        let turnsDir = workspacePath.appendingPathComponent("turns", isDirectory: true)
        guard let files = try? FileManager.default.contentsOfDirectory(at: turnsDir, includingPropertiesForKeys: nil, options: [.skipsHiddenFiles]) else {
            return []
        }
        var turns: [TurnRecord] = []
        for file in files where file.pathExtension == "md" {
            guard let raw = try? String(contentsOf: file, encoding: .utf8) else { continue }
            let parsed = FrontmatterParser.parse(raw)
            let fallbackID = file.deletingPathExtension().lastPathComponent
            let id = parsed.meta["id"] ?? fallbackID
            let seq = Int(parsed.meta["seq"] ?? "") ?? Int(fallbackID.dropFirst()) ?? 0
            let role = parsed.meta["role"] ?? "user"
            let provider = parsed.meta["provider"] ?? parsed.meta["author"] ?? ""
            let created = parsed.meta["created"] ?? ""
            let sha = parsed.meta["body_sha"] ?? ""
            turns.append(TurnRecord(
                id: id,
                seq: seq,
                role: role,
                provider: provider,
                created: created,
                bodySHA: sha,
                body: parsed.body
            ))
        }
        return turns.sorted { $0.seq < $1.seq }
    }

    static func loadDecisions(workspacePath: URL) -> [DecisionRecord] {
        let dir = workspacePath.appendingPathComponent("decisions", isDirectory: true)
        guard let files = try? FileManager.default.contentsOfDirectory(at: dir, includingPropertiesForKeys: nil, options: [.skipsHiddenFiles]) else {
            return []
        }
        var out: [DecisionRecord] = []
        for file in files where file.pathExtension == "md" {
            guard let raw = try? String(contentsOf: file, encoding: .utf8) else { continue }
            let parsed = FrontmatterParser.parse(raw)
            let id = parsed.meta["id"] ?? file.deletingPathExtension().lastPathComponent
            let status = parsed.meta["status"] ?? "active"
            let source = parsed.meta["source"] ?? ""
            let statement = parsed.meta["statement"] ?? parsed.body
            let quote = parsed.meta["quote"] ?? ""
            out.append(DecisionRecord(
                id: id,
                status: status,
                source: source,
                statement: statement,
                quote: quote
            ))
        }
        return out.sorted {
            if $0.status != $1.status { return $0.status == "active" }
            return $0.id < $1.id
        }
    }
}

enum CXRunner {
    static func detectBinary() -> String {
        let fm = FileManager.default
        let candidates = [
            // Relative to apps/macos/ContextlessAI.app/Contents/MacOS/ContextlessApp -> repo/bin/cx
            Bundle.main.bundleURL
                .deletingLastPathComponent()
                .deletingLastPathComponent()
                .deletingLastPathComponent()
                .appendingPathComponent("bin/cx").path,
            // Relative to current working directory
            URL(fileURLWithPath: fm.currentDirectoryPath).appendingPathComponent("bin/cx").path,
            // Common checkout & install locations
            fm.homeDirectoryForCurrentUser.appendingPathComponent("code/ContextlessAI/bin/cx").path,
            "/usr/local/bin/cx",
            "/opt/homebrew/bin/cx",
            fm.homeDirectoryForCurrentUser.appendingPathComponent(".local/bin/cx").path
        ]
        for path in candidates where fm.isExecutableFile(atPath: path) {
            return path
        }
        return "cx"
    }

    static func run(binary: String, arguments: [String]) async -> CommandResult {
        await withCheckedContinuation { continuation in
            DispatchQueue.global(qos: .userInitiated).async {
                let proc = Process()
                if binary.contains("/") {
                    proc.executableURL = URL(fileURLWithPath: binary)
                    proc.arguments = arguments
                } else {
                    proc.executableURL = URL(fileURLWithPath: "/usr/bin/env")
                    proc.arguments = [binary] + arguments
                }

                // Ensure standard Homebrew & Go paths are available when launched from Finder.
                var env = ProcessInfo.processInfo.environment
                let extraPaths = ["/opt/homebrew/bin", "/usr/local/bin", "\(FileManager.default.homeDirectoryForCurrentUser.path)/.local/bin", "\(FileManager.default.homeDirectoryForCurrentUser.path)/go/bin"]
                let currentPath = env["PATH"] ?? "/usr/bin:/bin:/usr/sbin:/sbin"
                env["PATH"] = (extraPaths + [currentPath]).joined(separator: ":")
                proc.environment = env

                let outPipe = Pipe()
                let errPipe = Pipe()
                proc.standardOutput = outPipe
                proc.standardError = errPipe

                do {
                    try proc.run()
                    let outData = outPipe.fileHandleForReading.readDataToEndOfFile()
                    let errData = errPipe.fileHandleForReading.readDataToEndOfFile()
                    proc.waitUntilExit()
                    let result = CommandResult(
                        exitCode: proc.terminationStatus,
                        stdout: String(data: outData, encoding: .utf8) ?? "",
                        stderr: String(data: errData, encoding: .utf8) ?? ""
                    )
                    continuation.resume(returning: result)
                } catch {
                    continuation.resume(returning: CommandResult(
                        exitCode: -1,
                        stdout: "",
                        stderr: "Failed to launch \(binary): \(error.localizedDescription)"
                    ))
                }
            }
        }
    }
}

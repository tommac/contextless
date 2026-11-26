// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "ContextlessApp",
    platforms: [
        .macOS(.v14)
    ],
    products: [
        .executable(name: "ContextlessApp", targets: ["ContextlessApp"])
    ],
    targets: [
        .executableTarget(
            name: "ContextlessApp",
            path: "Sources/ContextlessApp"
        )
    ]
)

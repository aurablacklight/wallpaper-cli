import Foundation
import SQLite3

@MainActor
class CLIDatabaseManager: ObservableObject {
    private var db: OpaquePointer?
    private let dbPath: String
    @Published var isConnected = false
    
    init() {
        let home = FileManager.default.homeDirectoryForCurrentUser
        self.dbPath = home.appendingPathComponent(".local/share/wallpaper-cli/wallpapers.db").path
    }
    
    func connect() async throws {
        guard !isConnected else { return }
        guard FileManager.default.fileExists(atPath: dbPath) else {
            throw NSError(domain: "CLIIntegration", code: -1, userInfo: [NSLocalizedDescriptionKey: "Database not found"])
        }
        let result = sqlite3_open(dbPath, &db)
        guard result == SQLITE_OK else { throw NSError(domain: "CLIIntegration", code: Int(result)) }
        isConnected = true
    }
    
    func getFavorites() async throws -> [CLIImageRecord] {
        try await connect()
        var images: [CLIImageRecord] = []
        let query = "SELECT i.* FROM favorites f JOIN images i ON f.image_hash = i.hash"
        var stmt: OpaquePointer?
        if sqlite3_prepare_v2(db, query, -1, &stmt, nil) == SQLITE_OK {
            while sqlite3_step(stmt) == SQLITE_ROW {
                images.append(parseImage(stmt))
            }
            sqlite3_finalize(stmt)
        }
        return images
    }
    
    func getPlaylists() async throws -> [CLIPlaylist] {
        try await connect()
        var playlists: [CLIPlaylist] = []
        let query = "SELECT id, name, description FROM playlists"
        var stmt: OpaquePointer?
        if sqlite3_prepare_v2(db, query, -1, &stmt, nil) == SQLITE_OK {
            while sqlite3_step(stmt) == SQLITE_ROW {
                playlists.append(CLIPlaylist(
                    id: String(cString: sqlite3_column_text(stmt, 0)!),
                    name: String(cString: sqlite3_column_text(stmt, 1)!),
                    description: sqlite3_column_text(stmt, 2).map { String(cString: $0) }
                ))
            }
            sqlite3_finalize(stmt)
        }
        return playlists
    }
    
    private func parseImage(_ stmt: OpaquePointer?) -> CLIImageRecord {
        CLIImageRecord(
            hash: String(cString: sqlite3_column_text(stmt, 0)!),
            source: String(cString: sqlite3_column_text(stmt, 1)!),
            localPath: String(cString: sqlite3_column_text(stmt, 4)!),
            rating: nil,
            isFavorite: true
        )
    }
}

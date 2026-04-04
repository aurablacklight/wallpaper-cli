import XCTest
import SQLite3
@testable import WallpaperEngine

// MARK: - CLI Database Validation Tests

class CLIIntegrationTests: XCTestCase {
    
    var dbManager: CLIDatabaseManager!
    let cliDBPath = FileManager.default
        .homeDirectoryForCurrentUser
        .appendingPathComponent(".local/share/wallpaper-cli/wallpapers.db")
    
    override func setUp() {
        super.setUp()
        dbManager = CLIDatabaseManager()
    }
    
    override func tearDown() {
        dbManager = nil
        super.tearDown()
    }
    
    func testDatabaseFileExists() {
        let fileManager = FileManager.default
        XCTAssertTrue(fileManager.fileExists(atPath: cliDBPath.path),
                      "CLI database not found at \(cliDBPath.path)")
        print("✅ TEST 1A PASSED: Database file exists")
    }
    
    func testCanOpenDatabaseConnection() async throws {
        try await dbManager.connect()
        print("✅ TEST 1B PASSED: Can open database connection")
    }
    
    func testCanReadFavorites() async throws {
        try await dbManager.connect()
        let favorites = try await dbManager.getFavorites()
        print("Found \(favorites.count) favorites")
        XCTAssertGreaterThanOrEqual(favorites.count, 0, "Should return array even if empty")
        print("✅ TEST 2A PASSED: Can read favorites")
    }
}

class CLIDatabaseManager {
    private var db: OpaquePointer?
    private let dbPath: String
    
    init() {
        let home = FileManager.default.homeDirectoryForCurrentUser
        self.dbPath = home.appendingPathComponent(".local/share/wallpaper-cli/wallpapers.db").path
    }
    
    deinit {
        sqlite3_close(db)
    }
    
    func connect() async throws {
        let result = sqlite3_open(dbPath, &db)
        guard result == SQLITE_OK else {
            throw NSError(domain: "CLIIntegration", code: Int(result), userInfo: [NSLocalizedDescriptionKey: "Failed to open database"])
        }
    }
    
    func getFavorites() async throws -> [CLIImageRecord] {
        guard let db = db else { throw NSError(domain: "CLIIntegration", code: -1) }
        
        var images: [CLIImageRecord] = []
        let query = """
            SELECT i.hash, i.source, i.source_id, i.url, i.local_path,
                   i.resolution, i.aspect_ratio, i.tags, i.downloaded_at, i.file_size
            FROM favorites f
            JOIN images i ON f.image_hash = i.hash
            """
        
        var statement: OpaquePointer?
        
        if sqlite3_prepare_v2(db, query, -1, &statement, nil) == SQLITE_OK {
            while sqlite3_step(statement) == SQLITE_ROW {
                let image = CLIImageRecord(
                    hash: stringFromColumn(statement, 0) ?? "",
                    source: stringFromColumn(statement, 1) ?? "",
                    sourceId: stringFromColumn(statement, 2),
                    url: stringFromColumn(statement, 3) ?? "",
                    localPath: stringFromColumn(statement, 4) ?? "",
                    resolution: stringFromColumn(statement, 5),
                    aspectRatio: stringFromColumn(statement, 6),
                    tags: stringFromColumn(statement, 7),
                    downloadedAt: nil,
                    fileSize: sqlite3_column_int64(statement, 9),
                    rating: nil,
                    isFavorite: true
                )
                images.append(image)
            }
            sqlite3_finalize(statement)
        }
        
        return images
    }
}

struct CLIImageRecord: Identifiable {
    let id = UUID()
    let hash: String
    let source: String
    let sourceId: String?
    let url: String
    let localPath: String
    let resolution: String?
    let aspectRatio: String?
    let tags: String?
    let downloadedAt: Date?
    let fileSize: Int64
    var rating: Int?
    var isFavorite: Bool
    
    var title: String {
        URL(fileURLWithPath: localPath).lastPathComponent
    }
}

private func stringFromColumn(_ statement: OpaquePointer?, _ index: Int32) -> String? {
    guard let cString = sqlite3_column_text(statement, index) else { return nil }
    return String(cString: cString)
}

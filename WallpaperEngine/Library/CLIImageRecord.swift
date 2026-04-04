import Foundation

struct CLIImageRecord: Identifiable {
    let id = UUID()
    let hash: String
    let source: String
    let localPath: String
    var rating: Int?
    var isFavorite: Bool
    
    var title: String {
        URL(fileURLWithPath: localPath).lastPathComponent
    }
    
    var fileURL: URL {
        URL(fileURLWithPath: localPath)
    }
}

struct CLIPlaylist: Identifiable {
    let id: String
    let name: String
    let description: String?
    var itemCount: Int = 0
}

enum CLICollectionType: String, CaseIterable, Identifiable {
    case favorites = "Favorites"
    case playlists = "Playlists"
    
    var id: String { rawValue }
    var icon: String {
        self == .favorites ? "star.fill" : "list.bullet.rectangle"
    }
}

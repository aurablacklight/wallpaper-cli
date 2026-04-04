import Foundation
import Combine

@MainActor
class CollectionsViewModel: ObservableObject {
    @Published var favorites: [CLIImageRecord] = []
    @Published var playlists: [CLIPlaylist] = []
    @Published var isLoading = false
    @Published var errorMessage: String?
    
    let dbManager = CLIDatabaseManager()
    
    func loadCollections() async {
        isLoading = true
        do {
            try await dbManager.connect()
            async let favTask = dbManager.getFavorites()
            async let playlistTask = dbManager.getPlaylists()
            favorites = try await favTask
            playlists = try await playlistTask
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}

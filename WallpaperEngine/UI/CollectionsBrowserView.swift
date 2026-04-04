import SwiftUI

struct CollectionsBrowserView: View {
    @StateObject private var viewModel = CollectionsViewModel()
    @EnvironmentObject var controller: WallpaperController
    @State private var selectedCollection: CLICollectionType = .favorites
    
    var body: some View {
        VStack {
            Picker("Collection", selection: $selectedCollection) {
                ForEach([CLICollectionType.favorites, .playlists]) { type in
                    Label(type.rawValue, systemImage: type.icon).tag(type)
                }
            }
            .pickerStyle(.segmented)
            .padding()
            
            if viewModel.isLoading {
                ProgressView()
            } else if selectedCollection == .favorites {
                favoritesGrid
            } else {
                playlistsGrid
            }
        }
        .task { await viewModel.loadCollections() }
    }
    
    private var favoritesGrid: some View {
        ScrollView {
            LazyVGrid(columns: [GridItem(.adaptive(minimum: 160))], spacing: 12) {
                ForEach(viewModel.favorites) { record in
                    Button(action: { setWallpaper(record) }) {
                        VStack {
                            Text(record.title).font(.caption)
                            if record.isFavorite {
                                Image(systemName: "star.fill").foregroundColor(.yellow)
                            }
                        }
                    }
                }
            }
            .padding()
        }
    }
    
    private var playlistsGrid: some View {
        List(viewModel.playlists) { playlist in
            Text(playlist.name)
        }
    }
    
    private func setWallpaper(_ record: CLIImageRecord) {
        controller.setWallpaper(CLIWallpaperAdapter.shared.makeWallpaperItem(from: record))
    }
}

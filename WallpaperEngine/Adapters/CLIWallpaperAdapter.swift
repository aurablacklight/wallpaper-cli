import Foundation
import AppKit

class CLIWallpaperAdapter {
    static let shared = CLIWallpaperAdapter()
    
    func makeWallpaperItem(from record: CLIImageRecord) -> WallpaperItem {
        let ext = record.fileURL.pathExtension.lowercased()
        let type: WallpaperType = ["mp4", "mov", "avi"].contains(ext) ? .video : .image
        
        return WallpaperItem(
            id: "cli-\(record.hash)",
            title: record.title,
            type: type,
            entryFile: record.fileURL,
            previewImage: record.fileURL,
            directoryURL: record.fileURL.deletingLastPathComponent(),
            isValid: FileManager.default.fileExists(atPath: record.localPath),
            source: "cli-\(record.source)"
        )
    }
}

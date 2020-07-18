package scanner

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/deluan/navidrome/utils"
)

type TagScanner struct {
	rootFolder string
	ds         model.DataStore
	detector   *changeDetector
	mapper     *mediaFileMapper
	firstRun   sync.Once
}

func NewTagScanner(rootFolder string, ds model.DataStore) *TagScanner {
	return &TagScanner{
		rootFolder: rootFolder,
		ds:         ds,
		detector:   newChangeDetector(rootFolder),
		mapper:     newMediaFileMapper(rootFolder),
		firstRun:   sync.Once{},
	}
}

type (
	artistMap map[string]struct{}
	albumMap  map[string]struct{}

	counters struct {
		added   int64
		updated int64
		deleted int64
	}
)

const (
	// filesBatchSize used for extract file metadata
	filesBatchSize = 100
)

// Scan algorithm overview:
// For each changed folder: Get all files from DB that starts with the folder, scan each file:
//	    if file in folder is newer, update the one in DB
//      if file in folder does not exists in DB, add
// 	    for each file in the DB that is not found in the folder, delete from DB
// For each deleted folder: delete all files from DB that starts with the folder path
// Only on first run, check if any folder under each changed folder is missing.
//      if it is, delete everything under it
// Create new albums/artists, update counters:
//      collect all albumIDs and artistIDs from previous steps
//	    refresh the collected albums and artists with the metadata from the mediafiles
// Delete all empty albums, delete all empty Artists
func (s *TagScanner) Scan(ctx context.Context, lastModifiedSince time.Time) error {
	start := time.Now()
	log.Trace(ctx, "Looking for changes in music folder", "folder", s.rootFolder)

	changed, deleted, err := s.detector.Scan(ctx, lastModifiedSince)
	if err != nil {
		return err
	}

	if len(changed)+len(deleted) == 0 {
		log.Debug(ctx, "No changes found in Music Folder", "folder", s.rootFolder)
		return nil
	}

	if log.CurrentLevel() >= log.LevelTrace {
		log.Info(ctx, "Folder changes found", "numChanged", len(changed), "numDeleted", len(deleted),
			"changed", strings.Join(changed, ";"), "deleted", strings.Join(deleted, ";"))
	} else {
		log.Info(ctx, "Folder changes found", "numChanged", len(changed), "numDeleted", len(deleted))
	}

	sort.Strings(changed)
	sort.Strings(deleted)

	updatedArtists := artistMap{}
	updatedAlbums := albumMap{}
	cnt := &counters{}

	for _, c := range changed {
		err := s.processChangedDir(ctx, c, updatedArtists, updatedAlbums, cnt)
		if err != nil {
			return err
		}
		// TODO Search for playlists and import (with `sync` on)
	}
	for _, c := range deleted {
		err := s.processDeletedDir(ctx, c, updatedArtists, updatedAlbums, cnt)
		if err != nil {
			return err
		}
		// TODO "Un-sync" all playlists synched from a deleted folder
	}

	err = s.flushAlbums(ctx, updatedAlbums)
	if err != nil {
		return err
	}

	err = s.flushArtists(ctx, updatedArtists)
	if err != nil {
		return err
	}

	s.firstRun.Do(func() {
		s.removeDeletedFolders(context.TODO(), changed, cnt)
	})

	err = s.ds.GC(log.NewContext(context.TODO()))
	log.Info("Finished Music Folder", "folder", s.rootFolder, "elapsed", time.Since(start),
		"added", cnt.added, "updated", cnt.updated, "deleted", cnt.deleted)

	return err
}

func (s *TagScanner) flushAlbums(ctx context.Context, updatedAlbums albumMap) error {
	if len(updatedAlbums) == 0 {
		return nil
	}
	var ids []string
	for id := range updatedAlbums {
		ids = append(ids, id)
		delete(updatedAlbums, id)
	}
	return s.ds.Album(ctx).Refresh(ids...)
}

func (s *TagScanner) flushArtists(ctx context.Context, updatedArtists artistMap) error {
	if len(updatedArtists) == 0 {
		return nil
	}
	var ids []string
	for id := range updatedArtists {
		ids = append(ids, id)
		delete(updatedArtists, id)
	}
	return s.ds.Artist(ctx).Refresh(ids...)
}

func (s *TagScanner) processChangedDir(ctx context.Context, dir string, updatedArtists artistMap, updatedAlbums albumMap, cnt *counters) error {
	dir = filepath.Join(s.rootFolder, dir)
	start := time.Now()

	// Load folder's current tracks from DB into a map
	currentTracks := map[string]model.MediaFile{}
	ct, err := s.ds.MediaFile(ctx).FindAllByPath(dir)
	if err != nil {
		return err
	}
	for _, t := range ct {
		currentTracks[t.Path] = t
	}

	// Load tracks FileInfo from the folder
	files, err := LoadAllAudioFiles(dir)
	if err != nil {
		return err
	}

	// If no files to process, return
	if len(files)+len(currentTracks) == 0 {
		return nil
	}

	// If track from folder is newer than the one in DB, select for update/insert in DB and delete from the current tracks
	log.Trace("Processing changed folder", "dir", dir, "tracksInDB", len(currentTracks), "tracksInFolder", len(files))
	var filesToUpdate []string
	for filePath, info := range files {
		c, ok := currentTracks[filePath]
		if !ok {
			filesToUpdate = append(filesToUpdate, filePath)
			cnt.added++
		}
		if ok && info.ModTime().After(c.UpdatedAt) {
			filesToUpdate = append(filesToUpdate, filePath)
			cnt.updated++
		}
		delete(currentTracks, filePath)

		// Force a refresh of the album and artist, to cater for cover art files. Ideally we would only do this
		// if there are any image file in the folder (TODO)
		err = s.updateAlbum(ctx, c.AlbumID, updatedAlbums)
		if err != nil {
			return err
		}
		err = s.updateArtist(ctx, c.AlbumArtistID, updatedArtists)
		if err != nil {
			return err
		}
	}

	numUpdatedTracks := 0
	numPurgedTracks := 0

	if len(filesToUpdate) > 0 {
		// Break the file list in chunks to avoid calling ffmpeg with too many parameters
		chunks := utils.BreakUpStringSlice(filesToUpdate, filesBatchSize)
		for _, chunk := range chunks {
			// Load tracks Metadata from the folder
			newTracks, err := s.loadTracks(chunk)
			if err != nil {
				return err
			}

			// If track from folder is newer than the one in DB, update/insert in DB
			log.Trace("Updating mediaFiles in DB", "dir", dir, "files", chunk, "numFiles", len(chunk))
			for i := range newTracks {
				n := newTracks[i]
				err := s.ds.MediaFile(ctx).Put(&n)
				if err != nil {
					return err
				}
				err = s.updateAlbum(ctx, n.AlbumID, updatedAlbums)
				if err != nil {
					return err
				}
				err = s.updateArtist(ctx, n.AlbumArtistID, updatedArtists)
				if err != nil {
					return err
				}
				numUpdatedTracks++
			}
		}
	}

	if len(currentTracks) > 0 {
		log.Trace("Deleting dangling tracks from DB", "dir", dir, "numTracks", len(currentTracks))
		// Remaining tracks from DB that are not in the folder are deleted
		for _, ct := range currentTracks {
			numPurgedTracks++
			err = s.updateAlbum(ctx, ct.AlbumID, updatedAlbums)
			if err != nil {
				return err
			}
			err = s.updateArtist(ctx, ct.AlbumArtistID, updatedArtists)
			if err != nil {
				return err
			}
			if err := s.ds.MediaFile(ctx).Delete(ct.ID); err != nil {
				return err
			}
			cnt.deleted++
		}
	}

	log.Info("Finished processing changed folder", "dir", dir, "updated", numUpdatedTracks, "purged", numPurgedTracks, "elapsed", time.Since(start))
	return nil
}

func (s *TagScanner) updateAlbum(ctx context.Context, albumId string, updatedAlbums albumMap) error {
	updatedAlbums[albumId] = struct{}{}
	if len(updatedAlbums) >= batchSize {
		err := s.flushAlbums(ctx, updatedAlbums)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TagScanner) updateArtist(ctx context.Context, artistId string, updatedArtists artistMap) error {
	updatedArtists[artistId] = struct{}{}
	if len(updatedArtists) >= batchSize {
		err := s.flushArtists(ctx, updatedArtists)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TagScanner) processDeletedDir(ctx context.Context, dir string, updatedArtists artistMap, updatedAlbums albumMap, cnt *counters) error {
	dir = filepath.Join(s.rootFolder, dir)
	start := time.Now()

	mfs, err := s.ds.MediaFile(ctx).FindAllByPath(dir)
	if err != nil {
		return err
	}
	for _, t := range mfs {
		err = s.updateAlbum(ctx, t.AlbumID, updatedAlbums)
		if err != nil {
			return err
		}
		err = s.updateArtist(ctx, t.AlbumArtistID, updatedArtists)
		if err != nil {
			return err
		}
	}

	log.Info("Finished processing deleted folder", "dir", dir, "purged", len(mfs), "elapsed", time.Since(start))
	c, err := s.ds.MediaFile(ctx).DeleteByPath(dir)
	cnt.deleted += c
	return err
}

func (s *TagScanner) removeDeletedFolders(ctx context.Context, changed []string, cnt *counters) {
	for _, dir := range changed {
		fullPath := filepath.Join(s.rootFolder, dir)
		paths, err := s.ds.MediaFile(ctx).FindPathsRecursively(fullPath)
		if err != nil {
			log.Error(ctx, "Error reading paths from DB", "path", dir, err)
			return
		}

		// If a path is unreadable, remove from the DB
		for _, path := range paths {
			if readable, err := utils.IsDirReadable(path); !readable {
				log.Info(ctx, "Path unavailable. Removing tracks from DB", "path", path, err)
				c, err := s.ds.MediaFile(ctx).DeleteByPath(path)
				if err != nil {
					log.Error(ctx, "Error removing MediaFiles from DB", "path", path, err)
				}
				cnt.deleted += c
			}
		}
	}
}

func (s *TagScanner) loadTracks(filePaths []string) (model.MediaFiles, error) {
	mds, err := ExtractAllMetadata(filePaths)
	if err != nil {
		return nil, err
	}

	var mfs model.MediaFiles
	for _, md := range mds {
		mf := s.mapper.toMediaFile(md)
		mfs = append(mfs, mf)
	}
	return mfs, nil
}

func LoadAllAudioFiles(dirPath string) (map[string]os.FileInfo, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}
	audioFiles := make(map[string]os.FileInfo)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		filePath := filepath.Join(dirPath, f.Name())
		if !utils.IsAudioFile(filePath) {
			continue
		}
		fi, err := os.Stat(filePath)
		if err != nil {
			log.Error("Could not stat file", "filePath", filePath, err)
		} else {
			audioFiles[filePath] = fi
		}
	}

	return audioFiles, nil
}

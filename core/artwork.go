package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strings"
	"time"

	"github.com/deluan/navidrome/core/cache"
	_ "golang.org/x/image/webp"

	"github.com/deluan/navidrome/conf"
	"github.com/deluan/navidrome/consts"
	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/deluan/navidrome/resources"
	"github.com/deluan/navidrome/utils"
	"github.com/dhowden/tag"
	"github.com/disintegration/imaging"
)

type Artwork interface {
	Get(ctx context.Context, id string, size int, out io.Writer) error
}

type ArtworkCache cache.FileCache

func NewArtwork(ds model.DataStore, cache ArtworkCache) Artwork {
	return &artwork{ds: ds, cache: cache}
}

type artwork struct {
	ds    model.DataStore
	cache cache.FileCache
}

type imageInfo struct {
	c          *artwork
	id         string
	path       string
	size       int
	lastUpdate time.Time
}

func (ci *imageInfo) String() string {
	return fmt.Sprintf("%s.%d.%s.%d", ci.path, ci.size, ci.lastUpdate.Format(time.RFC3339Nano), conf.Server.CoverJpegQuality)
}

func (a *artwork) Get(ctx context.Context, id string, size int, out io.Writer) error {
	path, lastUpdate, err := a.getImagePath(ctx, id)
	if err != nil && err != model.ErrNotFound {
		return err
	}

	info := &imageInfo{
		c:          a,
		id:         id,
		path:       path,
		size:       size,
		lastUpdate: lastUpdate,
	}

	r, err := a.cache.Get(ctx, info)
	if err != nil {
		log.Error(ctx, "Error accessing image cache", "path", path, "size", size, err)
		return err
	}
	defer r.Close()

	_, err = io.Copy(out, r)
	return err
}

func (a *artwork) getImagePath(ctx context.Context, id string) (path string, lastUpdated time.Time, err error) {
	// If id is an album cover ID
	if strings.HasPrefix(id, "al-") {
		log.Trace(ctx, "Looking for album art", "id", id)
		id = strings.TrimPrefix(id, "al-")
		var al *model.Album
		al, err = a.ds.Album(ctx).Get(id)
		if err != nil {
			return
		}
		if al.CoverArtId == "" {
			err = model.ErrNotFound
		}
		return al.CoverArtPath, al.UpdatedAt, err
	}

	log.Trace(ctx, "Looking for media file art", "id", id)

	// Check if id is a mediaFile cover id
	var mf *model.MediaFile
	mf, err = a.ds.MediaFile(ctx).Get(id)

	// If it is not, may be an albumId
	if err == model.ErrNotFound {
		return a.getImagePath(ctx, "al-"+id)
	}
	if err != nil {
		return
	}

	// If it is a mediaFile and it has cover art, return it
	if mf.HasCoverArt {
		return mf.Path, mf.UpdatedAt, nil
	}

	// if the mediaFile does not have a coverArt, fallback to the album cover
	log.Trace(ctx, "Media file does not contain art. Falling back to album art", "id", id, "albumId", "al-"+mf.AlbumID)
	return a.getImagePath(ctx, "al-"+mf.AlbumID)
}

func (a *artwork) getArtwork(ctx context.Context, id string, path string, size int) (reader io.Reader, err error) {
	defer func() {
		if err != nil {
			log.Warn(ctx, "Error extracting image", "path", path, "size", size, err)
			reader, err = resources.AssetFile().Open(consts.PlaceholderAlbumArt)
		}
	}()

	if path == "" {
		return nil, errors.New("empty path given for artwork")
	}

	var data []byte

	if size == 0 {
		// If requested original size, just read from the file
		if utils.IsAudioFile(path) {
			data, err = readFromTag(path)
		} else {
			data, err = readFromFile(path)
		}
	} else {
		// If requested a resized image, get the original (possibly from cache) and resize it
		a2 := NewArtwork(a.ds, a.cache)
		buf := new(bytes.Buffer)
		err = a2.Get(ctx, id, 0, buf)
		if err != nil {
			return
		}
		data, err = resizeImage(buf, size)
	}

	// Confirm the image is valid. Costly, but necessary
	_, _, err = image.Decode(bytes.NewReader(data))
	if err == nil {
		reader = bytes.NewReader(data)
	}

	return
}

func resizeImage(reader io.Reader, size int) ([]byte, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	m := imaging.Resize(img, size, size, imaging.Lanczos)
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, m, &jpeg.Options{Quality: conf.Server.CoverJpegQuality})
	return buf.Bytes(), err
}

func readFromTag(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	picture := m.Picture()
	if picture == nil {
		return nil, errors.New("file does not contain embedded art")
	}
	return picture.Data, nil
}

func readFromFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func NewImageCache() ArtworkCache {
	return cache.NewFileCache("Image", conf.Server.ImageCacheSize, consts.ImageCacheDir, consts.DefaultImageCacheMaxItems,
		func(ctx context.Context, arg fmt.Stringer) (io.Reader, error) {
			info := arg.(*imageInfo)
			reader, err := info.c.getArtwork(ctx, info.id, info.path, info.size)
			if err != nil {
				log.Error(ctx, "Error loading artwork art", "path", info.path, "size", info.size, err)
				return nil, err
			}
			return reader, nil
		})
}

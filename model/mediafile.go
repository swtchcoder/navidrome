package model

import (
	"mime"
	"time"
)

type MediaFile struct {
	ID          string
	Path        string
	Title       string
	Album       string
	Artist      string
	ArtistID    string
	AlbumArtist string
	AlbumID     string `parent:"album"`
	HasCoverArt bool
	TrackNumber int
	DiscNumber  int
	Year        int
	Size        string
	Suffix      string
	Duration    int
	BitRate     int
	Genre       string
	Compilation bool
	PlayCount   int
	PlayDate    time.Time
	Rating      int
	Starred     bool
	StarredAt   time.Time `idx:"Starred"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (mf *MediaFile) ContentType() string {
	return mime.TypeByExtension("." + mf.Suffix)
}

type MediaFiles []MediaFile

func (a MediaFiles) Len() int      { return len(a) }
func (a MediaFiles) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a MediaFiles) Less(i, j int) bool {
	return (a[i].DiscNumber*1000 + a[i].TrackNumber) < (a[j].DiscNumber*1000 + a[j].TrackNumber)
}

type MediaFileRepository interface {
	BaseRepository
	Put(m *MediaFile) error
	Get(id string) (*MediaFile, error)
	FindByAlbum(albumId string) (MediaFiles, error)
	GetStarred(options ...QueryOptions) (MediaFiles, error)
	PurgeInactive(active MediaFiles) error
	GetAllIds() ([]string, error)
	Search(q string, offset int, size int) (MediaFiles, error)
}

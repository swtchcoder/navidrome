package model

import "time"

type Album struct {
	Annotations

	ID                   string    `json:"id"            orm:"column(id)"`
	Name                 string    `json:"name"`
	CoverArtPath         string    `json:"coverArtPath"`
	CoverArtId           string    `json:"coverArtId"`
	ArtistID             string    `json:"artistId"      orm:"pk;column(artist_id)"`
	Artist               string    `json:"artist"`
	AlbumArtistID        string    `json:"albumArtistId" orm:"pk;column(album_artist_id)"`
	AlbumArtist          string    `json:"albumArtist"`
	MaxYear              int       `json:"maxYear"`
	MinYear              int       `json:"minYear"`
	Compilation          bool      `json:"compilation"`
	SongCount            int       `json:"songCount"`
	Duration             float32   `json:"duration"`
	Genre                string    `json:"genre"`
	FullText             string    `json:"fullText"`
	SortAlbumName        string    `json:"sortAlbumName"`
	SortArtistName       string    `json:"sortArtistName"`
	SortAlbumArtistName  string    `json:"sortAlbumArtistName"`
	OrderAlbumName       string    `json:"orderAlbumName"`
	OrderAlbumArtistName string    `json:"orderAlbumArtistName"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	Size                 int       `json:"size"`
}

type Albums []Album

type AlbumRepository interface {
	CountAll(...QueryOptions) (int64, error)
	Exists(id string) (bool, error)
	Get(id string) (*Album, error)
	FindByArtist(albumArtistId string) (Albums, error)
	GetAll(...QueryOptions) (Albums, error)
	GetRandom(...QueryOptions) (Albums, error)
	GetStarred(options ...QueryOptions) (Albums, error)
	Search(q string, offset int, size int) (Albums, error)
	Refresh(ids ...string) error
	AnnotatedRepository
}

func (a Album) GetAnnotations() Annotations {
	return a.Annotations
}

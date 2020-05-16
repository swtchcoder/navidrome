package persistence

import (
	"context"
	"strings"
	"time"

	. "github.com/Masterminds/squirrel"
	"github.com/astaxie/beego/orm"
	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/deluan/rest"
)

type playlist struct {
	ID        string `orm:"column(id)"`
	Name      string
	Comment   string
	Duration  float32
	Owner     string
	Public    bool
	Tracks    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type playlistRepository struct {
	sqlRepository
	sqlRestful
}

func NewPlaylistRepository(ctx context.Context, o orm.Ormer) model.PlaylistRepository {
	r := &playlistRepository{}
	r.ctx = ctx
	r.ormer = o
	r.tableName = "playlist"
	return r
}

func (r *playlistRepository) CountAll(options ...model.QueryOptions) (int64, error) {
	return r.count(Select(), options...)
}

func (r *playlistRepository) Exists(id string) (bool, error) {
	return r.exists(Select().Where(Eq{"id": id}))
}

func (r *playlistRepository) Delete(id string) error {
	return r.delete(Eq{"id": id})
}

func (r *playlistRepository) Put(p *model.Playlist) error {
	if p.ID == "" {
		p.CreatedAt = time.Now()
	}
	p.UpdatedAt = time.Now()
	pls := r.fromModel(p)
	_, err := r.put(pls.ID, pls)
	return err
}

func (r *playlistRepository) Get(id string) (*model.Playlist, error) {
	sel := r.newSelect().Columns("*").Where(Eq{"id": id})
	var res playlist
	err := r.queryOne(sel, &res)
	pls := r.toModel(&res)
	return &pls, err
}

func (r *playlistRepository) GetAll(options ...model.QueryOptions) (model.Playlists, error) {
	sel := r.newSelect(options...).Columns("*")
	var res []playlist
	err := r.queryAll(sel, &res)
	return r.toModels(res), err
}

func (r *playlistRepository) toModels(all []playlist) model.Playlists {
	result := make(model.Playlists, len(all))
	for i := range all {
		p := all[i]
		result[i] = r.toModel(&p)
	}
	return result
}

func (r *playlistRepository) toModel(p *playlist) model.Playlist {
	pls := model.Playlist{
		ID:        p.ID,
		Name:      p.Name,
		Comment:   p.Comment,
		Duration:  p.Duration,
		Owner:     p.Owner,
		Public:    p.Public,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if strings.TrimSpace(p.Tracks) != "" {
		tracks := strings.Split(p.Tracks, ",")
		for _, t := range tracks {
			pls.Tracks = append(pls.Tracks, model.MediaFile{ID: t})
		}
	}
	pls.Tracks = r.loadTracks(&pls)
	return pls
}

func (r *playlistRepository) fromModel(p *model.Playlist) playlist {
	pls := playlist{
		ID:        p.ID,
		Name:      p.Name,
		Comment:   p.Comment,
		Owner:     p.Owner,
		Public:    p.Public,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	// TODO Update duration with a SQL query, instead of loading all tracks
	p.Tracks = r.loadTracks(p)
	var newTracks []string
	for _, t := range p.Tracks {
		newTracks = append(newTracks, t.ID)
		pls.Duration += t.Duration
	}
	pls.Tracks = strings.Join(newTracks, ",")
	return pls
}

// TODO: Introduce a relation table for Playlist <-> MediaFiles, and rewrite this method in pure SQL
func (r *playlistRepository) loadTracks(p *model.Playlist) model.MediaFiles {
	if len(p.Tracks) == 0 {
		return nil
	}

	// Collect all ids
	ids := make([]string, len(p.Tracks))
	for i, t := range p.Tracks {
		ids[i] = t.ID
	}

	// Break the list in chunks, up to 50 items, to avoid hitting SQLITE_MAX_FUNCTION_ARG limit
	const chunkSize = 50
	var chunks [][]string
	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}

		chunks = append(chunks, ids[i:end])
	}

	// Query each chunk of media_file ids and store results in a map
	mfRepo := NewMediaFileRepository(r.ctx, r.ormer)
	trackMap := map[string]model.MediaFile{}
	for i := range chunks {
		idsFilter := Eq{"id": chunks[i]}
		tracks, err := mfRepo.GetAll(model.QueryOptions{Filters: idsFilter})
		if err != nil {
			log.Error(r.ctx, "Could not load playlist's tracks", "playlistName", p.Name, "playlistId", p.ID, err)
		}
		for _, t := range tracks {
			trackMap[t.ID] = t
		}
	}

	// Create a new list of tracks with the same order as the original
	newTracks := make(model.MediaFiles, len(p.Tracks))
	for i, t := range p.Tracks {
		newTracks[i] = trackMap[t.ID]
	}
	return newTracks
}

func (r *playlistRepository) Count(options ...rest.QueryOptions) (int64, error) {
	return r.CountAll(r.parseRestOptions(options...))
}

func (r *playlistRepository) Read(id string) (interface{}, error) {
	return r.Get(id)
}

func (r *playlistRepository) ReadAll(options ...rest.QueryOptions) (interface{}, error) {
	return r.GetAll(r.parseRestOptions(options...))
}

func (r *playlistRepository) EntityName() string {
	return "playlist"
}

func (r *playlistRepository) NewInstance() interface{} {
	return &model.Playlist{}
}

var _ model.PlaylistRepository = (*playlistRepository)(nil)
var _ model.ResourceRepository = (*playlistRepository)(nil)

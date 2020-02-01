package persistence

import (
	"context"
	"fmt"
	"sort"
	"strings"

	. "github.com/Masterminds/squirrel"
	"github.com/astaxie/beego/orm"
	"github.com/deluan/navidrome/conf"
	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/deluan/navidrome/utils"
	"github.com/deluan/rest"
)

type artistRepository struct {
	sqlRepository
	indexGroups utils.IndexGroups
}

func NewArtistRepository(ctx context.Context, o orm.Ormer) model.ArtistRepository {
	r := &artistRepository{}
	r.ctx = ctx
	r.ormer = o
	r.indexGroups = utils.ParseIndexGroups(conf.Server.IndexGroups)
	r.tableName = "artist"
	return r
}

func (r *artistRepository) selectArtist(options ...model.QueryOptions) SelectBuilder {
	return r.newSelectWithAnnotation(model.ArtistItemType, "id", options...).Columns("*")
}

func (r *artistRepository) CountAll(options ...model.QueryOptions) (int64, error) {
	return r.count(Select(), options...)
}

func (r *artistRepository) Exists(id string) (bool, error) {
	return r.exists(Select().Where(Eq{"id": id}))
}

func (r *artistRepository) getIndexKey(a *model.Artist) string {
	name := strings.ToLower(utils.NoArticle(a.Name))
	for k, v := range r.indexGroups {
		key := strings.ToLower(k)
		if strings.HasPrefix(name, key) {
			return v
		}
	}
	return "#"
}

func (r *artistRepository) Put(a *model.Artist) error {
	_, err := r.put(a.ID, a)
	if err != nil {
		return err
	}
	return r.index(a.ID, a.Name)
}

func (r *artistRepository) Get(id string) (*model.Artist, error) {
	sel := r.selectArtist().Where(Eq{"id": id})
	var res model.Artist
	err := r.queryOne(sel, &res)
	return &res, err
}

func (r *artistRepository) GetAll(options ...model.QueryOptions) (model.Artists, error) {
	sel := r.selectArtist(options...)
	var res model.Artists
	err := r.queryAll(sel, &res)
	return res, err
}

// TODO Cache the index (recalculate when there are changes to the DB)
func (r *artistRepository) GetIndex() (model.ArtistIndexes, error) {
	sq := r.selectArtist().OrderBy("name")
	var all model.Artists
	// TODO Paginate
	err := r.queryAll(sq, &all)
	if err != nil {
		return nil, err
	}

	fullIdx := make(map[string]*model.ArtistIndex)
	for _, a := range all {
		ax := r.getIndexKey(&a)
		idx, ok := fullIdx[ax]
		if !ok {
			idx = &model.ArtistIndex{ID: ax}
			fullIdx[ax] = idx
		}
		idx.Artists = append(idx.Artists, a)
	}
	var result model.ArtistIndexes
	for _, idx := range fullIdx {
		result = append(result, *idx)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (r *artistRepository) Refresh(ids ...string) error {
	type refreshArtist struct {
		model.Artist
		CurrentId   string
		AlbumArtist string
		Compilation bool
	}
	var artists []refreshArtist
	o := r.ormer
	sql := fmt.Sprintf(`
select f.artist_id as id,
       f.artist as name,
       f.album_artist,
       f.compilation,
       count(*) as album_count,
       a.id as current_id
from album f
         left outer join artist a on f.artist_id = a.id
where f.artist_id in ('%s') group by f.artist_id order by f.id`, strings.Join(ids, "','"))
	_, err := o.Raw(sql).QueryRows(&artists)
	if err != nil {
		return err
	}

	toInsert := 0
	toUpdate := 0
	for _, ar := range artists {
		if ar.Compilation {
			ar.AlbumArtist = "Various Artists"
		}
		if ar.AlbumArtist != "" {
			ar.Name = ar.AlbumArtist
		}
		if ar.CurrentId != "" {
			toUpdate++
		} else {
			toInsert++
		}
		err := r.Put(&ar.Artist)
		if err != nil {
			return err
		}
	}
	if toInsert > 0 {
		log.Debug(r.ctx, "Inserted new artists", "num", toInsert)
	}
	if toUpdate > 0 {
		log.Debug(r.ctx, "Updated artists", "num", toUpdate)
	}
	return err
}

func (r *artistRepository) GetStarred(options ...model.QueryOptions) (model.Artists, error) {
	sq := r.selectArtist(options...).Where("starred = true")
	var starred model.Artists
	err := r.queryAll(sq, &starred)
	return starred, err
}

func (r *artistRepository) PurgeEmpty() error {
	_, err := r.ormer.Raw("delete from artist where id not in (select distinct(artist_id) from album)").Exec()
	return err
}

func (r *artistRepository) Search(q string, offset int, size int) (model.Artists, error) {
	var results model.Artists
	err := r.doSearch(q, offset, size, &results, "name")
	return results, err
}

func (r *artistRepository) Count(options ...rest.QueryOptions) (int64, error) {
	return r.CountAll(r.parseRestOptions(options...))
}

func (r *artistRepository) Read(id string) (interface{}, error) {
	return r.Get(id)
}

func (r *artistRepository) ReadAll(options ...rest.QueryOptions) (interface{}, error) {
	return r.GetAll(r.parseRestOptions(options...))
}

func (r *artistRepository) EntityName() string {
	return "artist"
}

func (r *artistRepository) NewInstance() interface{} {
	return &model.Artist{}
}

var _ model.ArtistRepository = (*artistRepository)(nil)
var _ model.ArtistRepository = (*artistRepository)(nil)
var _ model.ResourceRepository = (*artistRepository)(nil)

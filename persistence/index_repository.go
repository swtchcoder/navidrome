package persistence

import (
	"sort"

	"github.com/astaxie/beego/orm"
	"github.com/cloudsonic/sonic-server/model"
)

type ArtistInfo struct {
	ID         int `orm:"pk;auto;column(id)"`
	Idx        string
	ArtistID   string `orm:"column(artist_id)"`
	Artist     string
	AlbumCount int
}

type artistIndexRepository struct {
	sqlRepository
}

func NewArtistIndexRepository() model.ArtistIndexRepository {
	r := &artistIndexRepository{}
	r.tableName = "artist_info"
	return r
}

func (r *artistIndexRepository) CountAll() (int64, error) {
	count := struct{ Count int64 }{}
	err := Db().Raw("select count(distinct(idx)) as count from artist_info").QueryRow(&count)
	if err != nil {
		return 0, err
	}
	return count.Count, nil
}

func (r *artistIndexRepository) Put(idx *model.ArtistIndex) error {
	return withTx(func(o orm.Ormer) error {
		_, err := r.newQuery(o).Filter("idx", idx.ID).Delete()
		if err != nil {
			return err
		}
		for _, artist := range idx.Artists {
			a := ArtistInfo{
				Idx:        idx.ID,
				ArtistID:   artist.ArtistID,
				Artist:     artist.Artist,
				AlbumCount: artist.AlbumCount,
			}
			err := r.insert(o, &a)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *artistIndexRepository) Get(id string) (*model.ArtistIndex, error) {
	var ais []ArtistInfo
	_, err := r.newQuery(Db()).Filter("idx", id).All(&ais)
	if err != nil {
		return nil, err
	}

	idx := &model.ArtistIndex{ID: id}
	idx.Artists = make([]model.ArtistInfo, len(ais))
	for i, a := range ais {
		idx.Artists[i] = model.ArtistInfo{
			ArtistID:   a.ArtistID,
			Artist:     a.Artist,
			AlbumCount: a.AlbumCount,
		}
	}
	return idx, err
}

func (r *artistIndexRepository) GetAll() (model.ArtistIndexes, error) {
	var all []ArtistInfo
	_, err := r.newQuery(Db()).OrderBy("idx", "artist").All(&all)
	if err != nil {
		return nil, err
	}

	fullIdx := make(map[string]*model.ArtistIndex)
	for _, a := range all {
		idx, ok := fullIdx[a.Idx]
		if !ok {
			idx = &model.ArtistIndex{ID: a.Idx}
			fullIdx[a.Idx] = idx
		}
		idx.Artists = append(idx.Artists, model.ArtistInfo{
			ArtistID:   a.ArtistID,
			Artist:     a.Artist,
			AlbumCount: a.AlbumCount,
		})
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

var _ model.ArtistIndexRepository = (*artistIndexRepository)(nil)

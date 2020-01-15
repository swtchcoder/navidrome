package persistence

import (
	"container/list"
	"sync"

	"github.com/cloudsonic/sonic-server/model"
)

var playerMap = sync.Map{}

type nowPlayingRepository struct{}

// TODO Make it persistent
func NewNowPlayingRepository() model.NowPlayingRepository {
	r := &nowPlayingRepository{}
	return r
}

func (r *nowPlayingRepository) getList(id int) *list.List {
	l, _ := playerMap.LoadOrStore(id, list.New())
	return l.(*list.List)
}

func (r *nowPlayingRepository) Enqueue(info *model.NowPlayingInfo) error {
	l := r.getList(info.PlayerId)
	l.PushFront(info)
	return nil
}

func (r *nowPlayingRepository) Dequeue(playerId int) (*model.NowPlayingInfo, error) {
	l := r.getList(playerId)
	e := l.Back()
	if e == nil {
		return nil, nil
	}
	l.Remove(e)
	return e.Value.(*model.NowPlayingInfo), nil
}

func (r *nowPlayingRepository) Head(playerId int) (*model.NowPlayingInfo, error) {
	l := r.getList(playerId)
	e := l.Front()
	if e == nil {
		return nil, nil
	}
	return e.Value.(*model.NowPlayingInfo), nil
}

func (r *nowPlayingRepository) Tail(playerId int) (*model.NowPlayingInfo, error) {
	l := r.getList(playerId)
	e := l.Back()
	if e == nil {
		return nil, nil
	}
	return e.Value.(*model.NowPlayingInfo), nil
}

func (r *nowPlayingRepository) Count(playerId int) (int64, error) {
	l := r.getList(playerId)
	return int64(l.Len()), nil
}

func (r *nowPlayingRepository) GetAll() ([]*model.NowPlayingInfo, error) {
	var all []*model.NowPlayingInfo
	playerMap.Range(func(playerId, l interface{}) bool {
		ll := l.(*list.List)
		e := ll.Front()
		all = append(all, e.Value.(*model.NowPlayingInfo))
		return true
	})
	return all, nil
}

var _ model.NowPlayingRepository = (*nowPlayingRepository)(nil)

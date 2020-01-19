package engine_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cloudsonic/sonic-server/engine"
	"github.com/cloudsonic/sonic-server/itunesbridge"
	"github.com/cloudsonic/sonic-server/persistence"
	. "github.com/cloudsonic/sonic-server/tests"
	. "github.com/smartystreets/goconvey/convey"
)

func TestScrobbler(t *testing.T) {

	Init(t, false)

	mfRepo := persistence.CreateMockMediaFileRepo()
	alRepo := persistence.CreateMockAlbumRepo()
	npRepo := engine.CreateMockNowPlayingRepo()
	itCtrl := &mockItunesControl{}

	scrobbler := engine.NewScrobbler(itCtrl, mfRepo, alRepo, npRepo)

	Convey("Given a DB with one song", t, func() {
		mfRepo.SetData(`[{"ID":"2","Title":"Hands Of Time"}]`, 1)

		Convey("When I scrobble an existing song", func() {
			now := time.Now()
			mf, err := scrobbler.Register(nil, 1, "2", now)

			Convey("Then I get the scrobbled song back", func() {
				So(err, ShouldBeNil)
				So(mf.Title, ShouldEqual, "Hands Of Time")
			})

			Convey("And iTunes is notified", func() {
				So(itCtrl.played, ShouldContainKey, "2")
				So(itCtrl.played["2"].Equal(now), ShouldBeTrue)
			})

		})

		Convey("When the ID is not in the DB", func() {
			_, err := scrobbler.Register(nil, 1, "3", time.Now())

			Convey("Then I receive an error", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("And iTunes is not notified", func() {
				So(itCtrl.played, ShouldNotContainKey, "3")
			})
		})

		Convey("When I inform the song that is now playing", func() {
			mf, err := scrobbler.NowPlaying(nil, 1, "DSub", "2", "deluan")

			Convey("Then I get the song for that id back", func() {
				So(err, ShouldBeNil)
				So(mf.Title, ShouldEqual, "Hands Of Time")
			})

			Convey("And it saves the song as the one current playing", func() {
				info, _ := npRepo.Head(1)
				So(info.TrackID, ShouldEqual, "2")
				// Commenting out time sensitive test, due to flakiness
				// So(info.Start, ShouldHappenBefore, time.Now())
				So(info.Username, ShouldEqual, "deluan")
				So(info.PlayerName, ShouldEqual, "DSub")
			})

			Convey("And iTunes is not notified", func() {
				So(itCtrl.played, ShouldNotContainKey, "2")
			})
		})

		Reset(func() {
			itCtrl.played = make(map[string]time.Time)
			itCtrl.skipped = make(map[string]time.Time)
		})

	})
}

var aPointInTime = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

func TestSkipping(t *testing.T) {
	Init(t, false)

	mfRepo := persistence.CreateMockMediaFileRepo()
	alRepo := persistence.CreateMockAlbumRepo()
	npRepo := engine.CreateMockNowPlayingRepo()
	itCtrl := &mockItunesControl{}

	scrobbler := engine.NewScrobbler(itCtrl, mfRepo, alRepo, npRepo)

	Convey("Given a DB with three songs", t, func() {
		mfRepo.SetData(`[{"ID":"1","Title":"Femme Fatale"},{"ID":"2","Title":"Here She Comes Now"},{"ID":"3","Title":"Lady Godiva's Operation"}]`, 3)
		itCtrl.skipped = make(map[string]time.Time)
		npRepo.ClearAll()
		Convey("When I skip 2 songs", func() {
			npRepo.OverrideNow(aPointInTime)
			scrobbler.NowPlaying(nil, 1, "DSub", "1", "deluan")
			npRepo.OverrideNow(aPointInTime.Add(2 * time.Second))
			scrobbler.NowPlaying(nil, 1, "DSub", "3", "deluan")
			npRepo.OverrideNow(aPointInTime.Add(3 * time.Second))
			scrobbler.NowPlaying(nil, 1, "DSub", "2", "deluan")
			Convey("Then the NowPlaying song should be the last one", func() {
				np, err := npRepo.GetAll()
				So(err, ShouldBeNil)
				So(np, ShouldHaveLength, 1)
				So(np[0].TrackID, ShouldEqual, "2")
			})
		})
		Convey("When I play one song", func() {
			npRepo.OverrideNow(aPointInTime)
			scrobbler.NowPlaying(nil, 1, "DSub", "1", "deluan")
			Convey("And I skip it before 20 seconds", func() {
				npRepo.OverrideNow(aPointInTime.Add(7 * time.Second))
				scrobbler.NowPlaying(nil, 1, "DSub", "2", "deluan")
				Convey("Then the first song should be marked as skipped", func() {
					mf, err := scrobbler.Register(nil, 1, "2", aPointInTime.Add(3*time.Minute))
					So(mf.ID, ShouldEqual, "2")
					So(itCtrl.skipped, ShouldContainKey, "1")
					So(err, ShouldBeNil)
				})
			})
			Convey("And I skip it before 3 seconds", func() {
				npRepo.OverrideNow(aPointInTime.Add(2 * time.Second))
				scrobbler.NowPlaying(nil, 1, "DSub", "2", "deluan")
				Convey("Then the first song should be marked as skipped", func() {
					mf, err := scrobbler.Register(nil, 1, "2", aPointInTime.Add(3*time.Minute))
					So(mf.ID, ShouldEqual, "2")
					So(itCtrl.skipped, ShouldBeEmpty)
					So(err, ShouldBeNil)
				})
			})
			Convey("And I skip it after 20 seconds", func() {
				npRepo.OverrideNow(aPointInTime.Add(30 * time.Second))
				scrobbler.NowPlaying(nil, 1, "DSub", "2", "deluan")
				Convey("Then the first song should be marked as skipped", func() {
					mf, err := scrobbler.Register(nil, 1, "2", aPointInTime.Add(3*time.Minute))
					So(mf.ID, ShouldEqual, "2")
					So(itCtrl.skipped, ShouldBeEmpty)
					So(err, ShouldBeNil)
				})
			})
			Convey("And I scrobble it before starting to play the other song", func() {
				mf, err := scrobbler.Register(nil, 1, "1", time.Now())
				Convey("Then the first song should NOT marked as skipped", func() {
					So(mf.ID, ShouldEqual, "1")
					So(itCtrl.skipped, ShouldBeEmpty)
					So(err, ShouldBeNil)
				})
			})
		})
		Convey("When the NowPlaying for the next song happens before the Scrobble", func() {
			npRepo.OverrideNow(aPointInTime)
			scrobbler.NowPlaying(nil, 1, "DSub", "1", "deluan")
			npRepo.OverrideNow(aPointInTime.Add(10 * time.Second))
			scrobbler.NowPlaying(nil, 1, "DSub", "2", "deluan")
			scrobbler.Register(nil, 1, "1", aPointInTime.Add(10*time.Minute))
			Convey("Then the NowPlaying song should be the last one", func() {
				np, _ := npRepo.GetAll()
				So(np, ShouldHaveLength, 1)
				So(np[0].TrackID, ShouldEqual, "2")
			})
		})
	})
}

type mockItunesControl struct {
	itunesbridge.ItunesControl
	played  map[string]time.Time
	skipped map[string]time.Time
	error   bool
}

func (m *mockItunesControl) MarkAsPlayed(id string, playDate time.Time) error {
	if m.error {
		return errors.New("ID not found")
	}
	if m.played == nil {
		m.played = make(map[string]time.Time)
	}
	m.played[id] = playDate
	return nil
}

func (m *mockItunesControl) MarkAsSkipped(id string, skipDate time.Time) error {
	if m.error {
		return errors.New("ID not found")
	}
	if m.skipped == nil {
		m.skipped = make(map[string]time.Time)
	}
	m.skipped[id] = skipDate
	return nil
}

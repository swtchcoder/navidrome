package agents

import (
	"context"
	"errors"

	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/tests"

	"github.com/navidrome/navidrome/conf"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agents", func() {
	var ctx context.Context
	var cancel context.CancelFunc
	var ds model.DataStore
	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		ds = &tests.MockDataStore{}
	})

	Describe("Placeholder", func() {
		var ag *Agents
		BeforeEach(func() {
			conf.Server.Agents = ""
			ag = New(ds)
		})

		It("calls the placeholder GetBiography", func() {
			Expect(ag.GetBiography(ctx, "123", "John Doe", "mb123")).To(Equal(placeholderBiography))
		})
		It("calls the placeholder GetImages", func() {
			images, err := ag.GetImages(ctx, "123", "John Doe", "mb123")
			Expect(err).ToNot(HaveOccurred())
			Expect(images).To(HaveLen(3))
			for _, i := range images {
				Expect(i.URL).To(BeElementOf(placeholderArtistImageSmallUrl, placeholderArtistImageMediumUrl, placeholderArtistImageLargeUrl))
			}
		})
	})

	Describe("Agents", func() {
		var ag *Agents
		var mock *mockAgent
		BeforeEach(func() {
			mock = &mockAgent{}
			Register("fake", func(ds model.DataStore) Interface {
				return mock
			})
			Register("empty", func(ds model.DataStore) Interface {
				return struct {
					Interface
				}{}
			})
			conf.Server.Agents = "empty,fake"
			ag = New(ds)
			Expect(ag.AgentName()).To(Equal("agents"))
		})

		Describe("GetMBID", func() {
			It("returns on first match", func() {
				Expect(ag.GetMBID(ctx, "123", "test")).To(Equal("mbid"))
				Expect(mock.Args).To(ConsistOf("123", "test"))
			})
			It("skips the agent if it returns an error", func() {
				mock.Err = errors.New("error")
				_, err := ag.GetMBID(ctx, "123", "test")
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(ConsistOf("123", "test"))
			})
			It("interrupts if the context is canceled", func() {
				cancel()
				_, err := ag.GetMBID(ctx, "123", "test")
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(BeEmpty())
			})
		})

		Describe("GetURL", func() {
			It("returns on first match", func() {
				Expect(ag.GetURL(ctx, "123", "test", "mb123")).To(Equal("url"))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123"))
			})
			It("skips the agent if it returns an error", func() {
				mock.Err = errors.New("error")
				_, err := ag.GetURL(ctx, "123", "test", "mb123")
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123"))
			})
			It("interrupts if the context is canceled", func() {
				cancel()
				_, err := ag.GetURL(ctx, "123", "test", "mb123")
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(BeEmpty())
			})
		})

		Describe("GetBiography", func() {
			It("returns on first match", func() {
				Expect(ag.GetBiography(ctx, "123", "test", "mb123")).To(Equal("bio"))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123"))
			})
			It("skips the agent if it returns an error", func() {
				mock.Err = errors.New("error")
				Expect(ag.GetBiography(ctx, "123", "test", "mb123")).To(Equal(placeholderBiography))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123"))
			})
			It("interrupts if the context is canceled", func() {
				cancel()
				_, err := ag.GetBiography(ctx, "123", "test", "mb123")
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(BeEmpty())
			})
		})

		Describe("GetImages", func() {
			It("returns on first match", func() {
				Expect(ag.GetImages(ctx, "123", "test", "mb123")).To(Equal([]ArtistImage{{
					URL:  "imageUrl",
					Size: 100,
				}}))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123"))
			})
			It("skips the agent if it returns an error", func() {
				mock.Err = errors.New("error")
				Expect(ag.GetImages(ctx, "123", "test", "mb123")).To(HaveLen(3))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123"))
			})
			It("interrupts if the context is canceled", func() {
				cancel()
				_, err := ag.GetImages(ctx, "123", "test", "mb123")
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(BeEmpty())
			})
		})

		Describe("GetSimilar", func() {
			It("returns on first match", func() {
				Expect(ag.GetSimilar(ctx, "123", "test", "mb123", 1)).To(Equal([]Artist{{
					Name: "Joe Dohn",
					MBID: "mbid321",
				}}))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123", 1))
			})
			It("skips the agent if it returns an error", func() {
				mock.Err = errors.New("error")
				_, err := ag.GetSimilar(ctx, "123", "test", "mb123", 1)
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123", 1))
			})
			It("interrupts if the context is canceled", func() {
				cancel()
				_, err := ag.GetSimilar(ctx, "123", "test", "mb123", 1)
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(BeEmpty())
			})
		})

		Describe("GetTopSongs", func() {
			It("returns on first match", func() {
				Expect(ag.GetTopSongs(ctx, "123", "test", "mb123", 2)).To(Equal([]Song{{
					Name: "A Song",
					MBID: "mbid444",
				}}))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123", 2))
			})
			It("skips the agent if it returns an error", func() {
				mock.Err = errors.New("error")
				_, err := ag.GetTopSongs(ctx, "123", "test", "mb123", 2)
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(ConsistOf("123", "test", "mb123", 2))
			})
			It("interrupts if the context is canceled", func() {
				cancel()
				_, err := ag.GetTopSongs(ctx, "123", "test", "mb123", 2)
				Expect(err).To(MatchError(ErrNotFound))
				Expect(mock.Args).To(BeEmpty())
			})
		})
	})
})

type mockAgent struct {
	Args []interface{}
	Err  error
}

func (a *mockAgent) AgentName() string {
	return "fake"
}

func (a *mockAgent) GetMBID(ctx context.Context, id string, name string) (string, error) {
	a.Args = []interface{}{id, name}
	if a.Err != nil {
		return "", a.Err
	}
	return "mbid", nil
}

func (a *mockAgent) GetURL(ctx context.Context, id, name, mbid string) (string, error) {
	a.Args = []interface{}{id, name, mbid}
	if a.Err != nil {
		return "", a.Err
	}
	return "url", nil
}

func (a *mockAgent) GetBiography(ctx context.Context, id, name, mbid string) (string, error) {
	a.Args = []interface{}{id, name, mbid}
	if a.Err != nil {
		return "", a.Err
	}
	return "bio", nil
}

func (a *mockAgent) GetImages(ctx context.Context, id, name, mbid string) ([]ArtistImage, error) {
	a.Args = []interface{}{id, name, mbid}
	if a.Err != nil {
		return nil, a.Err
	}
	return []ArtistImage{{
		URL:  "imageUrl",
		Size: 100,
	}}, nil
}

func (a *mockAgent) GetSimilar(ctx context.Context, id, name, mbid string, limit int) ([]Artist, error) {
	a.Args = []interface{}{id, name, mbid, limit}
	if a.Err != nil {
		return nil, a.Err
	}
	return []Artist{{
		Name: "Joe Dohn",
		MBID: "mbid321",
	}}, nil
}

func (a *mockAgent) GetTopSongs(ctx context.Context, id, artistName, mbid string, count int) ([]Song, error) {
	a.Args = []interface{}{id, artistName, mbid, count}
	if a.Err != nil {
		return nil, a.Err
	}
	return []Song{{
		Name: "A Song",
		MBID: "mbid444",
	}}, nil
}

package persistence

import (
	"sync"

	"github.com/cloudsonic/sonic-server/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NowPlayingRepository", func() {
	var repo model.NowPlayingRepository

	BeforeEach(func() {
		playerMap = sync.Map{}
		repo = NewNowPlayingRepository()
	})

	It("enqueues and dequeues records", func() {
		Expect(repo.Enqueue(&model.NowPlayingInfo{PlayerId: 1, TrackID: "AAA"})).To(BeNil())
		Expect(repo.Enqueue(&model.NowPlayingInfo{PlayerId: 1, TrackID: "BBB"})).To(BeNil())

		Expect(repo.Tail(1)).To(Equal(&model.NowPlayingInfo{PlayerId: 1, TrackID: "AAA"}))
		Expect(repo.Head(1)).To(Equal(&model.NowPlayingInfo{PlayerId: 1, TrackID: "BBB"}))

		Expect(repo.Count(1)).To(Equal(int64(2)))

		Expect(repo.Dequeue(1)).To(Equal(&model.NowPlayingInfo{PlayerId: 1, TrackID: "AAA"}))
		Expect(repo.Count(1)).To(Equal(int64(1)))
	})

	It("handles multiple players", func() {
		Expect(repo.Enqueue(&model.NowPlayingInfo{PlayerId: 1, TrackID: "AAA"})).To(BeNil())
		Expect(repo.Enqueue(&model.NowPlayingInfo{PlayerId: 1, TrackID: "BBB"})).To(BeNil())

		Expect(repo.Enqueue(&model.NowPlayingInfo{PlayerId: 2, TrackID: "CCC"})).To(BeNil())
		Expect(repo.Enqueue(&model.NowPlayingInfo{PlayerId: 2, TrackID: "DDD"})).To(BeNil())

		Expect(repo.GetAll()).To(ConsistOf([]*model.NowPlayingInfo{
			{PlayerId: 1, TrackID: "BBB"},
			{PlayerId: 2, TrackID: "DDD"},
		}))

		Expect(repo.Count(2)).To(Equal(int64(2)))
		Expect(repo.Count(2)).To(Equal(int64(2)))

		Expect(repo.Tail(1)).To(Equal(&model.NowPlayingInfo{PlayerId: 1, TrackID: "AAA"}))
		Expect(repo.Head(2)).To(Equal(&model.NowPlayingInfo{PlayerId: 2, TrackID: "DDD"}))
	})
})

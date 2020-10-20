package lastfm

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var httpClient *fakeHttpClient
	var client *Client

	BeforeEach(func() {
		httpClient = &fakeHttpClient{}
		client = NewClient("API_KEY", "pt", httpClient)
	})

	Describe("ArtistInfo", func() {
		It("returns an artist for a successful response", func() {
			f, _ := os.Open("tests/fixtures/lastfm.artist.getinfo.json")
			httpClient.res = http.Response{Body: f, StatusCode: 200}

			artist, err := client.ArtistGetInfo("U2")
			Expect(err).To(BeNil())
			Expect(artist.Name).To(Equal("U2"))
			Expect(httpClient.savedRequest.URL.String()).To(Equal(apiBaseUrl + "?api_key=API_KEY&artist=U2&format=json&lang=pt&method=artist.getInfo"))
		})

		It("fails if Last.FM returns an error", func() {
			httpClient.res = http.Response{
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{"error":3,"message":"Invalid Method - No method with that name in this package"}`)),
				StatusCode: 400,
			}

			_, err := client.ArtistGetInfo("U2")
			Expect(err).To(MatchError("last.fm error(3): Invalid Method - No method with that name in this package"))
		})

		It("fails if HttpClient.Do() returns error", func() {
			httpClient.err = errors.New("generic error")

			_, err := client.ArtistGetInfo("U2")
			Expect(err).To(MatchError("generic error"))
		})

		It("fails if returned body is not a valid JSON", func() {
			httpClient.res = http.Response{
				Body:       ioutil.NopCloser(bytes.NewBufferString(`<xml>NOT_VALID_JSON</xml>`)),
				StatusCode: 200,
			}

			_, err := client.ArtistGetInfo("U2")
			Expect(err).To(MatchError("invalid character '<' looking for beginning of value"))
		})

	})
})

type fakeHttpClient struct {
	res          http.Response
	err          error
	savedRequest *http.Request
}

func (c *fakeHttpClient) Do(req *http.Request) (*http.Response, error) {
	c.savedRequest = req
	if c.err != nil {
		return nil, c.err
	}
	return &c.res, nil
}

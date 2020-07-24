package core

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/deluan/navidrome/conf"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("File Caches", func() {
	BeforeEach(func() {
		conf.Server.DataFolder, _ = ioutil.TempDir("", "file_caches")
	})
	AfterEach(func() {
		os.RemoveAll(conf.Server.DataFolder)
	})

	Describe("NewFileCache", func() {
		It("creates the cache folder", func() {
			Expect(NewFileCache("test", "1k", "test", 0, nil)).ToNot(BeNil())

			_, err := os.Stat(filepath.Join(conf.Server.DataFolder, "test"))
			Expect(os.IsNotExist(err)).To(BeFalse())
		})

		It("creates the cache folder with invalid size", func() {
			fc, err := NewFileCache("test", "abc", "test", 0, nil)
			Expect(err).To(BeNil())
			Expect(fc.cache).ToNot(BeNil())
			Expect(fc.disabled).To(BeFalse())
		})

		It("returns empty if cache size is '0'", func() {
			fc, err := NewFileCache("test", "0", "test", 0, nil)
			Expect(err).To(BeNil())
			Expect(fc.cache).To(BeNil())
			Expect(fc.disabled).To(BeTrue())
		})
	})

	Describe("FileCache", func() {
		It("caches data if cache is enabled", func() {
			called := false
			fc, _ := NewFileCache("test", "1KB", "test", 0, func(ctx context.Context, arg fmt.Stringer) (io.Reader, error) {
				called = true
				return strings.NewReader(arg.String()), nil
			})
			// First call is a MISS
			s, err := fc.Get(context.TODO(), &testArg{"test"})
			Expect(err).To(BeNil())
			Expect(ioutil.ReadAll(s)).To(Equal([]byte("test")))

			// Second call is a HIT
			called = false
			s, err = fc.Get(context.TODO(), &testArg{"test"})
			Expect(err).To(BeNil())
			Expect(ioutil.ReadAll(s)).To(Equal([]byte("test")))
			Expect(called).To(BeFalse())
		})

		It("does not cache data if cache is disabled", func() {
			called := false
			fc, _ := NewFileCache("test", "0", "test", 0, func(ctx context.Context, arg fmt.Stringer) (io.Reader, error) {
				called = true
				return strings.NewReader(arg.String()), nil
			})
			// First call is a MISS
			s, err := fc.Get(context.TODO(), &testArg{"test"})
			Expect(err).To(BeNil())
			Expect(ioutil.ReadAll(s)).To(Equal([]byte("test")))

			// Second call is also a MISS
			called = false
			s, err = fc.Get(context.TODO(), &testArg{"test"})
			Expect(err).To(BeNil())
			Expect(ioutil.ReadAll(s)).To(Equal([]byte("test")))
			Expect(called).To(BeTrue())
		})
	})
})

type testArg struct{ s string }

func (t *testArg) String() string { return t.s }

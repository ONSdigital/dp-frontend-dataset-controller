package cache

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dp-topic-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewNavigationCache(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	Convey("Given a valid cache update interval which is greater than 0", t, func() {
		updateCacheInterval := 1 * time.Millisecond

		Convey("When NewNavigationCache is called", func() {
			testCache, err := NewNavigationCache(ctx, &updateCacheInterval)

			Convey("Then a navigation cache object should be successfully returned", func() {
				So(testCache, ShouldNotBeEmpty)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given no cache update interval (nil)", t, func() {
		Convey("When NewNavigationCache is called", func() {
			testCache, err := NewNavigationCache(ctx, nil)

			Convey("Then a cache object should be successfully returned", func() {
				So(testCache, ShouldNotBeEmpty)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})
}

func TestGetNavigationData(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockCacheList, err := GetMockCacheList(ctx, "en")
	if err != nil {
		t.Error("failed to get mock navigation cache list")
	}

	Convey("Given that navigation data exists in cache", t, func() {

		Convey("When GetNavigationData is called", func() {
			navigationData, err := mockCacheList.Navigation.GetNavigationData(ctx, "en")

			Convey("Then navigation cache data should be successfully returned", func() {
				So(navigationData, ShouldNotBeNil)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given the navigation data does not exist in cache", t, func() {
		mockCacheNavigation, err := NewNavigationCache(ctx, nil)
		So(err, ShouldBeNil)

		Convey("When GetNavigationData is called", func() {
			navigationData, err := mockCacheNavigation.GetNavigationData(ctx, "en")

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the navigation cache data returned should be empty", func() {
					So(navigationData, ShouldResemble, &models.Navigation{})
				})
			})
		})
	})
}

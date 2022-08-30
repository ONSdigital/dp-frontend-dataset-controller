package cache

import (
	"context"

	"github.com/ONSdigital/dp-topic-api/models"
)

// GetMockCacheList returns a mocked list of cache which contains the census topic cache and navigation cache
func GetMockCacheList(ctx context.Context, langs []string) (*CacheList, error) {

	testNavigationCache, err := getMockNavigationCache(ctx, langs)
	if err != nil {
		return nil, err
	}

	cacheList := CacheList{
		Navigation: testNavigationCache,
	}

	return &cacheList, nil
}

// getMockNavigationCache returns a mocked navigation cache which should have navigation data
func getMockNavigationCache(ctx context.Context, langs []string) (*NavigationCache, error) {
	testNavigationCache, err := NewNavigationCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	mockNavigationData := &models.Navigation{
		Description: "this is a test description",
	}

	for _, v := range langs {
		navigationlangKey := testNavigationCache.GetCachingKeyForNavigationLanguage(v)

		testNavigationCache.Set(navigationlangKey, mockNavigationData)
	}

	return testNavigationCache, nil
}

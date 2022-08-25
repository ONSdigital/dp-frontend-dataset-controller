package cache

import (
	"context"

	"github.com/ONSdigital/dp-topic-api/models"
)

// GetMockCacheList returns a mocked list of cache which contains the census topic cache and navigation cache
func GetMockCacheList(ctx context.Context, lang string) (*CacheList, error) {

	testNavigationCache, err := getMockNavigationCache(ctx, lang)
	if err != nil {
		return nil, err
	}

	cacheList := CacheList{
		Navigation: testNavigationCache,
	}

	return &cacheList, nil
}

// getMockNavigationCache returns a mocked navigation cache which should have navigation data
func getMockNavigationCache(ctx context.Context, lang string) (*NavigationCache, error) {
	testNavigationCache, err := NewNavigationCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	mockNavigationData := &models.Navigation{
		Description: "this is a test description",
	}

	navigationlangKey := testNavigationCache.GetCachingKeyForNavigationLanguage(lang)

	testNavigationCache.Set(navigationlangKey, mockNavigationData)

	return testNavigationCache, nil
}

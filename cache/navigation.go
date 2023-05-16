package cache

import (
	"context"
	"fmt"
	"time"

	dpcache "github.com/ONSdigital/dp-cache"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/log.go/v2/log"
)

// NavigationCache is a wrapper to dpcache.Cache which has additional fields and methods specifically for caching navigation data
type NavigationCache struct {
	*dpcache.Cache
}

// NewNavigationCache create a navigation cache object to be used in the service which will update at every updateInterval
// If updateInterval is nil, this means that the cache will only be updated once at the start of the service
func NewNavigationCache(ctx context.Context, updateInterval *time.Duration) (*NavigationCache, error) {
	config := dpcache.Config{
		UpdateInterval: updateInterval,
	}

	cache, err := dpcache.NewCache(ctx, config)
	if err != nil {
		logData := log.Data{
			"config": config,
		}
		log.Error(ctx, "failed to create cache from dpcache", err, logData)
		return nil, err
	}

	navigationCache := &NavigationCache{cache}

	return navigationCache, nil
}

// AddUpdateFunc adds an update function to the cache
func (nc *NavigationCache) AddUpdateFunc(key string, updateFunc func() *models.Navigation) {
	nc.UpdateFuncs[key] = func() (interface{}, error) {
		// error handled in updateFunc
		return updateFunc(), nil
	}
}

func (nc *NavigationCache) GetCachingKeyForNavigationLanguage(lang string) string {
	return fmt.Sprintf("%s___%s", NavigationCacheKey, lang)
}

func (nc *NavigationCache) GetNavigationData(ctx context.Context, lang string) (*models.Navigation, error) {
	key := nc.GetCachingKeyForNavigationLanguage(lang)
	navigationCacheInterface, ok := nc.Get(key)
	if !ok {
		err := fmt.Errorf("cached navigation data with key %s not found", key)
		log.Error(ctx, "failed to get cached navigation data", err)
		return &models.Navigation{}, err
	}

	navigationCache, _ := navigationCacheInterface.(*models.Navigation)
	return navigationCache, nil
}

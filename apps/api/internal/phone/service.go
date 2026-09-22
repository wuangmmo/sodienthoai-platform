package phone

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	Repository Repository
	Cache      *redis.Client
	TTL        time.Duration
}

func (s Service) Find(ctx context.Context, e164 string) (Number, error) {
	if s.Cache != nil {
		if raw, err := s.Cache.Get(ctx, cacheKey(e164)).Bytes(); err == nil {
			var n Number
			if json.Unmarshal(raw, &n) == nil {
				return n, nil
			}
		}
	}

	n, err := s.Repository.FindByE164(ctx, e164)
	if err != nil {
		return Number{}, err
	}

	if s.Cache != nil {
		if raw, err := json.Marshal(n); err == nil {
			ttl := s.TTL
			if ttl <= 0 { ttl = 10 * time.Minute }
			_ = s.Cache.Set(ctx, cacheKey(e164), raw, ttl).Err()
		}
	}
	return n, nil
}

func cacheKey(e164 string) string { return "phone:v1:" + e164 }

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

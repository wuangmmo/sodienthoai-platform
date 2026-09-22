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

func (s Service) invalidatePhoneCache(ctx context.Context, phoneID string) {
	if s.Cache == nil { return }
	if n, err := s.Repository.FindByID(ctx, phoneID); err == nil { _ = s.Cache.Del(ctx, cacheKey(n.E164)).Err() }
}

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

type LookupResult struct {
	Number     Number     `json:"number"`
	Identities []Identity `json:"identities"`
	Identified bool       `json:"identified"`
}

func (s Service) Lookup(ctx context.Context, e164 string) (LookupResult, error) {
	n, err := s.Find(ctx, e164)
	if err != nil {
		if IsNotFound(err) {
			country,calling,national := splitNormalized(e164)
			discovered,createErr := s.Repository.EnsureDiscovered(ctx,e164,country,calling,national)
			if createErr != nil { return LookupResult{}, createErr }
			_ = s.Repository.RecordLookup(ctx,e164,country,&discovered.ID,false)
			return LookupResult{Number:discovered,Identities:[]Identity{},Identified:false},nil
		}
		return LookupResult{}, err
	}
	identities, err := s.Repository.IdentitiesByPhoneID(ctx, n.ID)
	if err != nil { return LookupResult{}, err }
	_ = s.Repository.RecordLookup(ctx, e164, n.CountryCode, &n.ID, true)
	return LookupResult{Number:n, Identities:identities, Identified:len(identities)>0}, nil
}

func splitNormalized(e164 string)(country,calling,national string){
 if len(e164)>=3 && e164[:3]=="+84" { return "VN","84",e164[3:] }
 return "ZZ","",e164
}


func (s Service) Profile(ctx context.Context, e164 string) (Profile, error) {
	result, err := s.Lookup(ctx, e164)
	if err != nil { return Profile{}, err }
	signals, err := s.Repository.ProfileSignalsByPhoneID(ctx, result.Number.ID)
	if err != nil { return Profile{}, err }
	disputed, err := s.Repository.HasOpenIdentityDispute(ctx, result.Number.ID)
	if err != nil { return Profile{}, err }
	var primary *Identity
	for i := range result.Identities {
		if result.Identities[i].IsPrimary { primary = &result.Identities[i]; break }
	}
	return Profile{Number:result.Number,PrimaryIdentity:primary,Identities:result.Identities,Signals:signals,Identified:result.Identified,Disputed:disputed},nil
}

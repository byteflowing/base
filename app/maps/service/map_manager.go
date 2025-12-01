package service

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/byteflowing/base/app/maps/dal/query"
	"github.com/byteflowing/base/ecode"
	"github.com/byteflowing/base/pkg/redis"
	"github.com/byteflowing/base/pkg/utils/timex"
	enumsv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	mapsv1 "github.com/byteflowing/proto/gen/go/maps/v1"
)

//go:embed script_balancer_limit.lua
var scriptBalancerLimit string

//go:embed script_add.lua
var scriptBalancerAdd string

const (
	lockKeyPrefixFormat           = "%s:lock"
	balancerTargetFormat          = "balancer:%d:%d"  // balancer:source:itype
	interfaceTargetFormat         = "interface:%d:%d" // interface:source:itype
	interfaceStoreKeyPrefixFormat = "%s:%d:{%d}"      // prefix:source:itype
	interfaceStoreKeyFormat       = "%s:%d:{%d}:%d"   // prefix:source:itype:map_id
	balancerKeyFormat             = "%s:%d:{%d}"      // prefix:source:itype
	quotaKeyPrefixFormat          = "%s:%d:{%d}"      // prefix:source:itype
	limiterKeyPrefixFormat        = "%s:%d:{%d}"      // prefix:source:itype
)

// MapManager 统一管理地图服务的负载均衡和限流功能
type MapManager struct {
	rdb  *redis.Redis
	db   *query.Query
	lock *redis.Lock

	storePrefix       string
	limitPrefix       string
	balancerKeyPrefix string
}

// NewMapManager 创建新的地图管理器
func NewMapManager(
	prefix string,
	rdb *redis.Redis,
	db *query.Query,
) *MapManager {
	return &MapManager{
		rdb: rdb,
		db:  db,
		lock: redis.NewLock(rdb, &redis.LockOption{
			Prefix: fmt.Sprintf(lockKeyPrefixFormat, prefix),
			Tries:  3,
			TTL:    30 * time.Second,
			Wait:   10 * time.Millisecond,
		}),
		storePrefix:       prefix + ":" + "store",
		limitPrefix:       prefix + ":" + "limit",
		balancerKeyPrefix: prefix + ":" + "balancer",
	}
}

// GetInterfaceSummaryWithLimit 获取接口摘要并检查限制
// 该方法将负载均衡和限流检查合并到一个Redis操作中，减少Redis请求次数
func (m *MapManager) GetInterfaceSummaryWithLimit(
	ctx context.Context,
	source enumsv1.MapSource,
	iType enumsv1.MapInterfaceType,
	request int64,
	mapID *int64,
) (*mapsv1.MapInterfaceSummary, error) {
	// Prepare keys and arguments for the Lua script
	balancerKey := m.getBalancerKey(source, iType)
	storeKeyPrefix := m.getInterfaceStoreKeyPrefix(source, iType)
	dailyQuotaKeyPrefix := m.getDailyQuotaKeyPrefix(source, iType)
	rateLimitKeyPrefix := m.getRateLimiterKeyPrefix(source, iType)

	var mapIDValue int64
	if mapID != nil {
		mapIDValue = *mapID
	}
	keys := []string{balancerKey, storeKeyPrefix, dailyQuotaKeyPrefix, rateLimitKeyPrefix}
	args := []interface{}{
		mapIDValue,
		request,
	}
	// Execute the Lua script
	result, err := m.rdb.Eval(ctx, scriptBalancerLimit, keys, args...).Result()
	if err != nil {
		errMsg := err.Error()
		// balancer set not load
		if errMsg == "MAP_SET_NOT_EXISTS" {
			if err := m.loadBalancerFromDB(ctx, balancerKey, source, iType); err != nil {
				return nil, err
			}
			// Retry the Lua script
			return m.GetInterfaceSummaryWithLimit(ctx, source, iType, request, mapID)
		}

		// no resource
		if errMsg == "NO_RESOURCE" {
			return nil, ecode.ErrMapsNoResource
		}

		if strings.HasPrefix(errMsg, "INTERFACE_NOT_EXISTS:") {
			mapIDStr := strings.TrimPrefix(errMsg, "INTERFACE_NOT_EXISTS:")
			mid, parseErr := strconv.ParseInt(mapIDStr, 10, 64)
			if parseErr != nil {
				return nil, parseErr
			}
			if err := m.loadInterfaceFromDB(ctx, source, iType, mid); err != nil {
				return nil, err
			}
			// Retry the Lua script
			return m.GetInterfaceSummaryWithLimit(ctx, source, iType, request, mapID)
		}

		// Handle specific error cases
		if strings.HasPrefix(errMsg, "DAILY_LIMIT_EXCEEDED:") {
			mapIDStr := strings.TrimPrefix(errMsg, "INTERFACE_NOT_EXISTS:")
			mid, parseErr := strconv.ParseInt(mapIDStr, 10, 64)
			if parseErr != nil {
				return nil, parseErr
			}
			if err := m.RemoveMapInterface(ctx, source, iType, mid); err != nil {
				return nil, err
			}
			return m.GetInterfaceSummaryWithLimit(ctx, source, iType, request, mapID)
		}

		// Handle rate limit wait
		if strings.HasPrefix(errMsg, "RATE_LIMIT_WAIT:") {
			// Extract wait time from error message
			waitMsStr := strings.TrimPrefix(err.Error(), "RATE_LIMIT_WAIT:")
			waitMs, parseErr := strconv.ParseInt(waitMsStr, 10, 64)
			if parseErr != nil {
				return nil, parseErr
			}
			// Wait for the specified time
			waitDuration := time.Duration(waitMs) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitDuration):
				// After waiting, retry the operation
				return m.GetInterfaceSummaryWithLimit(ctx, source, iType, request, mapID)
			}
		}
		return nil, err
	}

	// Parse the result - it's a serialized MapInterfaceSummary
	bytes, ok := result.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid redis lua response")
	}

	// Unmarshal the MapInterfaceSummary
	summary := &mapsv1.MapInterfaceSummary{}
	if err := proto.Unmarshal(bytes, summary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map interface summary: %w", err)
	}

	return summary, nil
}

func (m *MapManager) AddMapInterface(ctx context.Context, source enumsv1.MapSource, iType enumsv1.MapInterfaceType, mapID int64) error {
	balancerKey := m.getBalancerKey(source, iType)
	return m.rdb.Eval(ctx, scriptBalancerAdd, []string{balancerKey}, mapID).Err()
}

func (m *MapManager) RemoveMapInterface(ctx context.Context, source enumsv1.MapSource, iType enumsv1.MapInterfaceType, mapID int64) error {
	balancerKey := m.getBalancerKey(source, iType)
	interfaceKey := m.getInterfaceKey(source, iType, mapID)
	pipe := m.rdb.Pipeline()
	pipe.HDel(ctx, balancerKey, strconv.FormatInt(mapID, 10))
	pipe.Del(ctx, interfaceKey)
	_, err := pipe.Exec(ctx)
	return err
}

func (m *MapManager) DeleteMapInterfaceCache(ctx context.Context, source enumsv1.MapSource, iType enumsv1.MapInterfaceType, mapID int64) error {
	key := m.getInterfaceKey(source, iType, mapID)
	return m.rdb.Del(ctx, key).Err()
}

// loadBalancerFromDB 从数据库加载负载均衡数据
func (m *MapManager) loadBalancerFromDB(ctx context.Context, key string, source enumsv1.MapSource, iType enumsv1.MapInterfaceType) error {
	target := m.getBalancerKey(source, iType)
	identifier, err := m.lock.Acquire(ctx, target)
	if err != nil {
		return err
	}
	defer m.lock.Release(ctx, target, identifier)
	res, err := m.rdb.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if res > 0 {
		return nil
	}
	q := m.db.MapAccount
	mapAccounts, err := q.WithContext(ctx).Where(
		q.MapSource.Eq(int16(source)),
		q.OwnerType.Eq(int16(enumsv1.MapOwnerType_MAP_OWNER_TYPE_SELF)),
		q.Status.Eq(int16(enumsv1.MapStatus_MAP_STATUS_OK)),
	).Select(q.ID).Find()
	if err != nil {
		return err
	}
	mapIDs := make([]int64, 0, len(mapAccounts))
	for _, m := range mapAccounts {
		mapIDs = append(mapIDs, m.ID)
	}
	iQ := m.db.MapInterface
	interfaces, err := iQ.WithContext(ctx).Where(
		iQ.InterfaceType.Eq(int32(iType)),
		iQ.MapID.In(mapIDs...),
		iQ.DailyLimit.Gt(0),
		iQ.SecondLimit.Gt(0),
	).Find()
	if err != nil {
		return err
	}
	ids := make([]interface{}, 0, len(interfaces))
	for _, i := range interfaces {
		ids = append(ids, i.MapID)
	}
	if len(ids) > 0 {
		if err := m.rdb.SAdd(ctx, key, ids...).Err(); err != nil {
			return err
		}
		mills := timex.EndOfDayMillis()
		ttl := time.Duration(mills) * time.Millisecond
		return m.rdb.Expire(ctx, key, ttl).Err()
	}
	return ecode.ErrNoResource
}

// loadInterfaceFromDB 从数据库加载存储数据
func (m *MapManager) loadInterfaceFromDB(ctx context.Context, source enumsv1.MapSource, iType enumsv1.MapInterfaceType, mapID int64) error {
	target := m.getInterfaceTarget(source, iType)
	identifier, err := m.lock.Acquire(ctx, target)
	if err != nil {
		return err
	}
	defer m.lock.Release(ctx, target, identifier)
	q := m.db.MapAccount
	account, err := q.WithContext(ctx).Where(q.ID.Eq(mapID)).First()
	if err != nil {
		return err
	}
	iQ := m.db.MapInterface
	is, err := iQ.WithContext(ctx).Where(
		iQ.MapID.Eq(mapID),
		iQ.InterfaceType.Eq(int32(iType)),
		iQ.DailyLimit.Gt(0),
		iQ.DailyLimit.Gt(0),
	).First()
	if err != nil {
		return err
	}
	ttl := m.getInterfaceTTL(enumsv1.MapOwnerType(account.OwnerType))
	storeMap := map[string]interface{}{
		"daily_limit":    is.DailyLimit,
		"second_limit":   is.SecondLimit,
		"key":            account.Key,
		"map_id":         mapID,
		"source":         source,
		"interface_type": iType,
	}
	interfaceStoreKey := m.getInterfaceKey(source, iType, mapID)
	if err := m.rdb.HMSet(ctx, interfaceStoreKey, storeMap).Err(); err != nil {
		return err
	}
	if err := m.rdb.Expire(ctx, interfaceStoreKey, ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (m *MapManager) getInterfaceStoreKeyPrefix(source enumsv1.MapSource, iType enumsv1.MapInterfaceType) string {
	return fmt.Sprintf(interfaceStoreKeyPrefixFormat, m.storePrefix, source, iType)
}

func (m *MapManager) getDailyQuotaKeyPrefix(source enumsv1.MapSource, iType enumsv1.MapInterfaceType) string {
	return fmt.Sprintf(quotaKeyPrefixFormat, m.storePrefix, source, iType)
}

func (m *MapManager) getRateLimiterKeyPrefix(source enumsv1.MapSource, iType enumsv1.MapInterfaceType) string {
	return fmt.Sprintf(limiterKeyPrefixFormat, m.storePrefix, source, iType)
}

func (m *MapManager) getBalancerKey(source enumsv1.MapSource, iType enumsv1.MapInterfaceType) string {
	return fmt.Sprintf(balancerKeyFormat, m.balancerKeyPrefix, source, iType)
}

func (m *MapManager) getInterfaceKey(source enumsv1.MapSource, iType enumsv1.MapInterfaceType, mapID int64) string {
	return fmt.Sprintf(interfaceStoreKeyFormat, m.storePrefix, source, iType, mapID)
}

func (m *MapManager) getBalancerTarget(source enumsv1.MapSource, iType enumsv1.MapInterfaceType) string {
	return fmt.Sprintf(balancerTargetFormat, source, iType)
}

func (m *MapManager) getInterfaceTarget(source enumsv1.MapSource, iType enumsv1.MapInterfaceType) string {
	return fmt.Sprintf(interfaceTargetFormat, source, iType)
}

// getInterfaceTTL 获取接口TTL
func (m *MapManager) getInterfaceTTL(t enumsv1.MapOwnerType) time.Duration {
	if t == enumsv1.MapOwnerType_MAP_OWNER_TYPE_SELF {
		return 168 * time.Hour
	}
	return 24 * time.Hour
}

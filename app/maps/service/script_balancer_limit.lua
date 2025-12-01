-- KEYS[1] = set key (for balancer)
-- KEYS[2] = store key prefix (for loading map info)
-- KEYS[3] = daily quota key prefix (for limiter)
-- KEYS[4] = rate limit key prefix (for limiter)
-- ARGV[1] = map id (when provided, skip balancer)
-- ARGV[2] = request count

local map_id = ARGV[1]
local request_count = tonumber(ARGV[2]) or 1
if request_count <= 0 then
    request_count = 1
end

if not map_id or map_id == "" then
    if redis.call("EXISTS", KEYS[1]) == 0 then
        return {err="MAP_SET_NOT_EXISTS"}
    end
    local nodes = redis.call("SRANDMEMBER", KEYS[1], 1)
    if not nodes or #nodes == 0 then
        return {err="NO_RESOURCE"}
    end
    map_id = nodes[1]
end

-- Load map interface info from Redis cache
local store_key = KEYS[2] .. ":" .. map_id
local cached_info = redis.call("GET", store_key)
if not cached_info then
    return {err="INTERFACE_NOT_EXISTS:" .. map_id}
end

local store_data = redis.call("HGETALL", store_key)

local daily_limit = 0
local second_limit = 0

if store_data and #store_data > 0 then
    for i = 1, #store_data, 2 do
        local field = store_data[i]
        local value = store_data[i + 1]
        if field == "daily_limit" then
            daily_limit = tonumber(value)
        elseif field == "second_limit" then
            second_limit = tonumber(value)
        end
    end
end

-- ======================
-- Daily limit
-- ======================
if daily_limit > 0 then
    local daily_quota_key = KEYS[3] .. ":" .. map_id
    local val = redis.call("GET", daily_quota_key)
    local current = val and tonumber(val) or 0

    if current + request_count > daily_limit then
        return {err="DAILY_LIMIT_EXCEEDED:".. map_id}
    end

    local newv = redis.call("INCRBY", daily_quota_key, request_count)

    if tonumber(newv) == request_count then
        local now = redis.call("TIME")
        local ts = tonumber(now[1])
        local ttl = 86400 - (ts % 86400)
        redis.call("EXPIRE", daily_quota_key, ttl)
    end
end

-- ======================
-- Token bucket without float
-- ======================
if second_limit > 0 then
    local rate_limit_key = KEYS[4] .. ":" .. map_id

    -- 使用整数缩放避免浮点
    local SCALE = 1000
    local capacity_scaled = second_limit * SCALE
    local need_scaled = request_count * SCALE
    local rate_per_ms_scaled = second_limit  -- scaled 后每 ms 加 second_limit

    local now = redis.call("TIME")
    local now_ms = now[1] * 1000 + math.floor(now[2] / 1000)

    local data = redis.call("HMGET", rate_limit_key, "tokens", "ts")
    local tokens_scaled = tonumber(data[1])
    local ts = tonumber(data[2])

    if not tokens_scaled or not ts then
        tokens_scaled = capacity_scaled
        ts = now_ms
    end

    -- refill
    local delta_ms = now_ms - ts
    if delta_ms < 0 then delta_ms = 0 end

    tokens_scaled = tokens_scaled + delta_ms * rate_per_ms_scaled
    if tokens_scaled > capacity_scaled then
        tokens_scaled = capacity_scaled
    end

    if tokens_scaled >= need_scaled then
        tokens_scaled = tokens_scaled - need_scaled
        redis.call("HMSET", rate_limit_key, "tokens", tokens_scaled, "ts", now_ms)
        redis.call("PEXPIRE", rate_limit_key, 2000)
    else
        local need = need_scaled - tokens_scaled
        local wait_ms = math.ceil(need / rate_per_ms_scaled)

        -- 写回当前状态
        redis.call("HMSET", rate_limit_key, "tokens", tokens_scaled, "ts", now_ms)
        redis.call("PEXPIRE", rate_limit_key, 2000)

        return {err="RATE_LIMIT_WAIT:" .. wait_ms}
    end
end

return cached_info
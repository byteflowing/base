-- KEYS[1] = 限流 key
-- ARGV[1] = capacity (桶最大容量)
-- ARGV[2] = interval_ms (补充满桶的时间间隔，毫秒)
-- ARGV[3] = requested (本次请求需要的 token 数)
-- 返回: { allowed (1/0), remaining_tokens, wait_ms }

local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local interval_ms = tonumber(ARGV[2])
local requested = tonumber(ARGV[3])

-- 计算速率：每毫秒增加多少个 token
local rate = capacity / interval_ms

-- Redis 时间（毫秒级）
local now = redis.call("TIME")
local now_ms = now[1] * 1000 + math.floor(now[2] / 1000)

-- 获取当前桶状态
local data = redis.call("HMGET", key, "tokens", "ts")
local tokens = tonumber(data[1])
local ts = tonumber(data[2])

if tokens == nil then
    tokens = capacity
    ts = now_ms
end

-- 计算经过时间，补充 token
local delta_ms = now_ms - ts
if delta_ms < 0 then
    delta_ms = 0
end

local added = delta_ms * rate
tokens = math.min(capacity, tokens + added)

local allowed = 0
local wait_ms = 0

if tokens >= requested then
    allowed = 1
    tokens = tokens - requested
else
    wait_ms = math.ceil((requested - tokens) / rate)
end

-- 更新桶状态
redis.call("HMSET", key, "tokens", tokens, "ts", now_ms)

-- TTL = interval 的 2 倍，保证桶在空闲时仍能自我过期
redis.call("PEXPIRE", key, interval_ms * 2)

return { allowed, tostring(tokens), wait_ms }
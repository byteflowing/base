-- KEYS[1] = 限流 key
-- ARGV[1] = 窗口时间(毫秒)
-- ARGV[2] = 配额上限
-- ARGV[3] = 一次消耗的数量
-- 返回值 {0/1, 当前值, 剩余ttl} 0：拒绝 1：允许

local ttl = tonumber(ARGV[1])
local quota = tonumber(ARGV[2])
local n = tonumber(ARGV[3])

-- 当前值
local val = redis.call("GET", KEYS[1])
local current = 0
if val then
    current = tonumber(val)
end

-- 超限检查
if current + n > quota then
    return {0, current, redis.call("PTTL", KEYS[1])}
end

-- 增加计数
current = redis.call("INCRBY", KEYS[1], n)

-- 首次创建 key 设置 TTL
if current == n then
    redis.call("PEXPIRE", KEYS[1], ttl)
    return {1, current, ttl}
end

return {1, current, redis.call("PTTL", KEYS[1])}
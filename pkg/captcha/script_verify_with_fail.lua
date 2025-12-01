-- KEYS[1] = 主key（存储目标值）
-- KEYS[2] = 失败计数key
-- ARGV[1] = 期望的值
-- ARGV[2] = 最大失败次数

local value = redis.call("GET", KEYS[1])
if not value then
    return {-1, 0}  -- 主key不存在或已过期
end

local maxFails = tonumber(ARGV[2])
local fails = tonumber(redis.call("GET", KEYS[2]) or "0")

if fails >= maxFails then
    redis.call("DEL", KEYS[1], KEYS[2])
    return {-2, fails}  -- 达到最大失败次数
end

if value ~= ARGV[1] then
    fails = redis.call("INCR", KEYS[2])
    if fails == 1 then
        local ttl = redis.call("TTL", KEYS[1])
        if ttl > 0 then
            redis.call("EXPIRE", KEYS[2], ttl)
        end
    end
    return {-3, fails}  -- 值不匹配
end

redis.call("DEL", KEYS[1], KEYS[2])
return {1, 0}  -- 成功
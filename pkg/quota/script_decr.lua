-- KEYS[1] = 限流 key
-- ARGV[1] = 要回退的数量
-- 返回值: 回退后的当前值（key 不存在时返回 0）

local n = tonumber(ARGV[1])
local val = redis.call("GET", KEYS[1])
if not val then
    return 0
end

local current = tonumber(val)
local newVal = current - n
if newVal < 0 then
    newVal = 0
end

redis.call("SET", KEYS[1], newVal, "KEEPTTL")
return newVal
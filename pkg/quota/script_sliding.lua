-- 多 key 固定窗口限流脚本
-- KEYS = {k1, k2, ..., kn}
-- ARGV = {duration1, limit1, duration2, limit2, ...}
-- 返回: {1/0, ttl, index}
-- 其中 1通过，0 不通过， ttl 剩余时间， index 表示第几个 key 触发了限制；0 表示都未超限

for i = 1, #KEYS do
    local key = KEYS[i]
    local duration = tonumber(ARGV[(i - 1) * 2 + 1])
    local quota = tonumber(ARGV[(i - 1) * 2 + 2])

    local val = redis.call("GET", key)
    if val and tonumber(val) >= quota then
        return {0, redis.call("PTTL", key), i-1, quota}
    end

    local current = redis.call("INCR", key)
    if current == 1 then
        redis.call("PEXPIRE", key, duration)
    end
end

return {1, 0, 0, current}
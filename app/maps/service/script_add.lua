-- KEYS[1] = set key
-- ARGV = 成员列表
if #ARGV == 0 then
    return 0
end
if redis.call('EXISTS', KEYS[1]) == 1 then
    return redis.call('SADD', KEYS[1], unpack(ARGV))
else
    return 0
end
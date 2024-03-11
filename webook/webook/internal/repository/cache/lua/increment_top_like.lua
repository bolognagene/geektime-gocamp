local key = KEYS[1]
-- +1 或者 -1
local delta = tonumber(ARGV[2])
-- 对应到的是 ZSet 中的 member
local member = ARGV[3]
local limit = ARGV[4]
local count = redis.call("ZCARD", key)
local score = redis.call("ZSCORE", key, member)
-- 该member不存在 ， 那么只有在ZSET个数小于limit 且 增加1 时才加入ZSET
if score ~= nil then
    if delta == 1 and count < limit then
        redis.call("ZADD", key, delta, member)
        return 0
    end
else
    redis.call("ZADD", key, delta, member)
    return 0
end

return 1
local key = KEYS[1]
-- +1 或者 -1
local delta = tonumber(ARGV[1])
-- 对应到的是 ZSet 中的 member
local member = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local count = redis.call("ZCARD", key)
local score = redis.call("ZSCORE", key, member)
-- zset不存在，直接返回
if count == 0 then
    return 4
end
-- 该member不存在 ， 那么只有在ZSET个数小于limit 且 增加1 时才加入ZSET
if score == nil then
    if delta == 1 and count < limit then
        redis.call("ZINCRBY", key, delta, member)
        return 2
    end
else
    redis.call("ZINCRBY", key, delta, member)
    return 3
end

return 1
-- zset 的 key
local key = KEYS[1]
local expireKey = key..":expire"
-- member
local member = ARGV[1]
-- current time
local now = tonumber(ARGV[2])

local exprired_at = tonumber(redis.call("ZSCORE", expireKey, member))
-- 过期了
if exprired_at < now then
    redis.call("ZREM", key, member)
    redis.call("ZREM", expireKey, member)
    return -1
else
    return redis.call("ZSCORE", key, member)
end

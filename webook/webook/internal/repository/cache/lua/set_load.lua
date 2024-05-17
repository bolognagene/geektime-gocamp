-- zset 的 key
local key = KEYS[1]
local expireKey = key..":expire"
-- member
local member = ARGV[1]
-- load
local load = tonumber(ARGV[2])
-- 过期时间
local exprired_at = tonumber(ARGV[3])

redis.call("ZADD", key, load, member)
redis.call("ZADD", expireKey, exprired_at, member)
return 0


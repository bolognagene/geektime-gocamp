-- zset 的 key
local key = KEYS[1]
local expireKey = key..":expire"
-- N
local n = tonumber(ARGV[1])
-- current time
local now = tonumber(ARGV[2])
-- type, 0 is low load, 1 is high load
local type = tonumber(ARGV[3])
-- 返回的值
local members

-- Remove expired key
local expired_members = redis.call('ZRANGEBYSCORE', expireKey, '-inf', now)
for _, member in ipairs(expired_members) do
    redis.call('ZREM', key, member)
    redis.call('ZREM', expireKey, member)
end

if type == 0 then
    members = redis.call("ZRANGE", key, 0, n)
else
    members = redis.call("ZRANGE", key, 0, n, "REV")
end

return members

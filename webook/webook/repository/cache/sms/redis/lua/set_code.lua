--你的验证码在 Redis 上的 key
-- phone_code:login:152xxxxxxxx
local key = KEYS[1]
-- 验证次数，默认为3
local cnt = 3
-- 传进来的验证码
local code = ARGV[1]
-- 过期时间
local ttl = tonumber(redis.call("ttl", key))
if ttl == -1 then
    --    key 存在，但是没有过期时间
    -- 系统错误，你的同事手贱，手动设置了这个 key，但是没给过期时间
    return -2
    --    540 = 600-60 九分钟
elseif ttl == -2 or ttl < 540 then
    redis.call('hset', key, 'count', cnt, 'code', code)
    redis.call('expire', key, 600)
    -- 完美，符合预期
    return 0
else
    -- 发送太频繁
    return -1
end


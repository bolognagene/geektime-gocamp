local key = KEYS[1]
-- 用户输入的 code
local inputCode = ARGV[1]
local cnt = tonumber(redis.call('hget', key, 'count'))
local code = redis.call('hget', key, 'code')
if cnt <= 0 then
--    说明，用户一直输错，有人搞你
--    或者已经用过了，也是有人搞你
--    -1  -> ErrCodeVerifyTooManyTimes
    return -1
elseif inputCode == code then
-- 输入对了
-- 用完，不能再用了
    redis.call('hset', key, 'cnt', -1)
    return 0
else
-- 用户手一抖，输错了
-- 可验证次数 -1
-- -2 -> ErrUnknownForCode
    cnt = cnt-1
    redis.call("set", key, 'cnt', cnt)
    return -2
end
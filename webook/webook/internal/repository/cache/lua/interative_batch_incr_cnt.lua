local keys = KEYS[1]
-- 对应到的是 hincrby 中的 field
local cntKey = ARGV[1]
-- +1 或者 -1
local delta = tonumber(ARGV[2])
-- 遍历keys，处理每个key
for i, key in ipairs(keys) do
    local exists = redis.call("EXISTS", key)
    if exists == 1 or delta > 0 then
        redis.call("HINCRBY", key, cntKey, delta)
        -- 说明自增成功了
    end
end

return 0
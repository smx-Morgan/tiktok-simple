-- 检测是否是预期值
--- 如果是 删除，如果不是返回一个值
if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end
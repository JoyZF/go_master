local viewLimit = redis.call("GET", "page:" .. KEYS[1] .. ":viewLimit")
local currentViews = redis.call("GET", "page:" .. KEYS[1] .. ":views")
-- convert them to a number. redis.call("GET") returns a string
viewLimit = tonumber(viewLimit)
currentViews = tonumber(currentViews)
-- the viewing limit has been reached
if currentViews >= viewLimit then
    return "no"
end
-- the user can view the page, let's add a new view to the views key
redis.log(redis.LOG_WARNING, currentViews)
redis.call("SET", "page:" .. KEYS[1] .. ":views", currentViews + 1)
return "yes"
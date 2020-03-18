require "zhelpers"
local zmq = require "lzmq"
local json = require "json"
-- sudo luarocks install uuid
local socket = require("socket")  -- gettime() has higher precision than os.time()
local uuid = require("uuid")
uuid.seed()

local MY_ID_STR = "1999"
local RPC_URL = "10.0.0.22" -- "localhost" -- "10.0.0.22"
local server_listening = false
local init_ping_id = uuid()
local SUBSCRIPTION_ID_LEN = #init_ping_id
subscription_ids = {}

-- Prepare our context and publisher
print("precontext")
local context = zmq.context()
print("postcontext")
print("tcp://" .. RPC_URL .. ":5570")
local client, err = context:socket(zmq.DEALER, {
    identity = MY_ID_STR;
    connect = "tcp://" .. RPC_URL .. ":5570";
})
print("post socket")
zassert(client, err)
print("post assert")

local msg = ".....PING" .. MY_ID_STR .. init_ping_id
print("about to send message: " .. msg)
local err2 = client:send_multipart{msg}
print("send!")
print(err2)

-- listen for PING response, blocking
local flags = 0
local raw_msg = client:recv_multipart()
print("received!")
print(raw_msg)
print(raw_msg[0]) -- nil
print(raw_msg[1]) -- "1999bd096d5b-e5bb-4425-8c8a-3d109f53a2641584412138839"

-- TODO: prehook

-- TODO: subcribe to stuff
function subscribe(query_strings, callback)
    subscription_id = uuid()
    query = {id = subscription_id, facts = query_strings}
    query_msg = json.encode(query)
    subscription_ids[subscription_id] = callback
    msg = "SUBSCRIBE" .. MY_ID_STR .. query_msg
    print(msg)
    local err_sub = client:send_multipart{msg}
    print("send sub!")
    print(err_sub)
end

function sub_callback(results)
    print("INSIDE CALLBACK!")
    print(results)
end

subscribe({"$ $ I am a turtle card"}, sub_callback)

-- TODO: listen loop
function listen(blocking)
    flags = 0
    if blocking == false then
        flags = 1 -- zmq.NOBLOCK
    end
    local raw_msg = client:recv_multipart(flags)
    print("received!")
    print(raw_msg)
    if raw_msg ~= nil then
        
        if #raw_msg > 0 then
            print(raw_msg[1])
            -- 1999b4a4f075-9fde-41ae-c1cb-bea4ea0b2b381584545958956[{}]
            local source_len = 4
            local server_send_time_len = 13
            local id = raw_msg:sub(source_len, source_len + SUBSCRIPTION_ID_LEN)
            local val = raw_msg:sub(source_len + SUBSCRIPTION_ID_LEN + server_send_time_len)
            print("ID")
            print(id)
            print("VAL")
            print(val)
            if id == init_ping_id then
                server_listening = true
            elseif subscription_ids[id] ~= nil then
                print("Found matching sub id")
                local callback = subscription_ids[id]
                callback(val)
            end
        end
    end
    print("TODO")
end

function sleep(sec)
    socket.select(nil, nil, sec)
end

while true do
    listen(false)
    sleep(0.5)
end

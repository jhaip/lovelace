require "zhelpers"
local zmq = require "lzmq"

local MY_ID_STR = "1999"
local RPC_URL = "10.0.0.22" -- "localhost" -- "10.0.0.22"
local init_ping_id = "bd096d5b-e5bb-4425-8c8a-3d109f53a264"

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
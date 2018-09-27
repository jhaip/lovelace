import time
import zmq
import logging
import json
import uuid
import random

logging.basicConfig(level=logging.INFO)

context = zmq.Context()
rpc_url = "localhost"
sub_socket = context.socket(zmq.SUB)
sub_socket.connect("tcp://{0}:5556".format(rpc_url))
pub_socket = context.socket(zmq.PUB)
pub_socket.connect("tcp://{0}:5555".format(rpc_url))

MY_ID = 1
MY_ID_STR = str(MY_ID).zfill(4)
SUBSCRIPTION_ID_LEN = len(str(uuid.uuid4()))
init_ping_id = str(uuid.uuid4())
server_listening = False
select_ids = {}
subscription_ids = {}

def claim(fact_string):
    pub_socket.send_string("....CLAIM{}{}".format(MY_ID_STR, fact_string), zmq.NOBLOCK)

def retract(fact_string):
    pub_socket.send_string("..RETRACT{}{}".format(MY_ID_STR, fact_string), zmq.NOBLOCK)

def select(query_strings, callback):
    select_id = str(uuid.uuid4())
    query = {
        "id": select_id,
        "facts": query_strings
    }
    query_msg = json.dumps(query)
    select_ids[select_id] = callback
    msg = "...SELECT{}{}".format(MY_ID_STR, query_msg)
    logging.info(msg)
    pub_socket.send_string(msg, zmq.NOBLOCK)

def subscribe(query_strings, callback):
    subscription_id = str(uuid.uuid4())
    query = {
        "id": subscription_id,
        "facts": query_strings
    }
    query_msg = json.dumps(query)
    subscription_ids[subscription_id] = callback
    msg = "SUBSCRIBE{}{}".format(MY_ID_STR, query_msg)
    logging.info(msg)
    pub_socket.send_string(msg, zmq.NOBLOCK)

def listen():
    global server_listening
    while True:
        try:
            string = sub_socket.recv_string(flags=zmq.NOBLOCK)
            source_len = 4
            id = string[source_len:(source_len + SUBSCRIPTION_ID_LEN)]
            val = string[(source_len + SUBSCRIPTION_ID_LEN):]
            if id == init_ping_id:
                server_listening = True
                return
            if id in select_ids:
                callback = select_ids[id]
                del select_ids[id]
                callback(val)
            elif id in subscription_ids:
                logging.info(string)
                callback = subscription_ids[id]
                callback(val)
            # else:
            #     logging.info("UNRECOGNIZED:")
            #     logging.info(string)
        except zmq.Again:
            break
    # logging.info("loop")
    # time.sleep(0.01)

def init(my_id, prehook=None, selects=[], subscriptions=[]):
    global MY_ID, MY_ID_STR
    MY_ID = my_id
    MY_ID_STR = str(MY_ID).zfill(4)
    sub_socket.setsockopt_string(zmq.SUBSCRIBE, MY_ID_STR)
    start = time.time()
    while not server_listening:
        pub_socket.send_string(".....PING{}{}".format(MY_ID_STR, init_ping_id), zmq.NOBLOCK)
        listen()
    end = time.time()
    print("INIT TIME: {} ms".format((end - start)*1000.0))
    # time.sleep(0.2)
    if prehook:
        prehook()
    for s in selects:
        query = s[0]
        callback = s[1]
        select(query, callback)
    for s in subscriptions:
        query = s[0]
        callback = s[1]
        subscribe(query, callback)
    while True:
        listen()

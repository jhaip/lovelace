import time
import zmq
import logging
import json
import uuid
import random
import os
import sys

context = zmq.Context()
rpc_url = "localhost"
sub_socket = context.socket(zmq.SUB)
sub_socket.connect("tcp://{0}:5555".format(rpc_url))
pub_socket = context.socket(zmq.PUB)
pub_socket.connect("tcp://{0}:5556".format(rpc_url))

MY_ID = None
MY_ID_STR = None
SUBSCRIPTION_ID_LEN = len(str(uuid.uuid4()))
init_ping_id = str(uuid.uuid4())
server_listening = False
select_ids = {}
subscription_ids = {}

py_subscriptions = []
py_prehook = None

def get_my_id_str():
    global MY_ID_STR
    return MY_ID_STR


def claim(fact_string):
    pub_socket.send_string("....CLAIM{}{}".format(
        MY_ID_STR, fact_string), zmq.NOBLOCK)


def batch(batch_claims):
    pub_socket.send_string("....BATCH{}{}".format(
        MY_ID_STR, json.dumps(batch_claims)), zmq.NOBLOCK)


def retract(fact_string):
    pub_socket.send_string("..RETRACT{}{}".format(
        MY_ID_STR, fact_string), zmq.NOBLOCK)


def select(query_strings, callback):
    select_id = str(uuid.uuid4())
    query = {
        "id": select_id,
        "facts": query_strings
    }
    query_msg = json.dumps(query)
    select_ids[select_id] = callback
    msg = "...SELECT{}{}".format(MY_ID_STR, query_msg)
    logging.debug(msg)
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
    logging.debug(msg)
    pub_socket.send_string(msg, zmq.NOBLOCK)


def parse_results(val):
    json_val = json.loads(val)
    results = []
    for result in json_val:
        new_result = {}
        for key in result:
            value_type = result[key][0]
            if value_type == "integer":
                new_result[key] = int(result[key][1])
            elif value_type == "float":
                new_result[key] = float(result[key][1])
            else:
                new_result[key] = str(result[key][1])
        results.append(new_result)
    return results


def listen():
    global server_listening
    while True:
        try:
            string = sub_socket.recv_string(flags=zmq.NOBLOCK)
            source_len = 4
            server_send_time_len = 13
            id = string[source_len:(source_len + SUBSCRIPTION_ID_LEN)]
            val = string[(source_len + SUBSCRIPTION_ID_LEN +
                          server_send_time_len):]
            if id == init_ping_id:
                server_listening = True
                return
            if id in select_ids:
                callback = select_ids[id]
                del select_ids[id]
                callback(val)
            elif id in subscription_ids:
                logging.debug(string)
                callback = subscription_ids[id]
                callback(parse_results(val))
            # else:
            #     logging.info("UNRECOGNIZED:")
            #     logging.info(string)
        except zmq.Again:
            break
    # logging.info("loop")
    time.sleep(0.01)


def init(root_filename, skipListening=False):
    global MY_ID, MY_ID_STR, py_subscriptions, py_prehook
    scriptName = os.path.basename(root_filename)
    scriptNameNoExtension = os.path.splitext(scriptName)[0]
    fileDir = os.path.dirname(os.path.realpath(root_filename))
    logPath = os.path.join(fileDir, 'logs/' + scriptNameNoExtension + '.log')
    logging.basicConfig(filename=logPath, level=logging.INFO)
    MY_ID = (scriptName.split(".")[0]).split("__")[0]
    MY_ID_STR = str(MY_ID).zfill(4)
    print("INSIDE INIT:")
    print(MY_ID)
    print(MY_ID_STR)
    print(logPath)
    print("-")
    sub_socket.setsockopt_string(zmq.SUBSCRIBE, MY_ID_STR)
    start = time.time()
    while not server_listening:
        pub_socket.send_string(".....PING{}{}".format(
            MY_ID_STR, init_ping_id), zmq.NOBLOCK)
        listen()
    end = time.time()
    print("INIT TIME: {} ms".format((end - start)*1000.0))
    logging.info("INIT TIME: {} ms".format((end - start)*1000.0))
    # time.sleep(0.2)
    if py_prehook:
        py_prehook()
    # for s in selects:
    #     query = s[0]
    #     callback = s[1]
    #     select(query, callback)
    for s in py_subscriptions:
        query = s[0]
        callback = s[1]
        subscribe(query, callback)
    if skipListening:
        return
    while True:
        listen()


def prehook(func):
    global py_prehook
    py_prehook = func

    def function_wrapper(x):
        func(x)
    return function_wrapper


def subscription(expr):
    def subscription_decorator(func):
        global py_subscriptions
        py_subscriptions.append((expr, func))

        def function_wrapper(x):
            func(x)
        return function_wrapper
    return subscription_decorator

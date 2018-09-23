from parser import parse
from server import claim_fact
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

sub_socket.setsockopt_string(zmq.SUBSCRIBE, "CLAIM[PROGRAM/")

# Receive claims, retracts, selects, subscribes

def claim(fact_string, source):
    fact = parse(fact_string)
    claim_fact(fact, source)

def retract(fact_string):
    fact = parse(fact_string)
    retract_fact(fact)

# pub_socket.send_string(s, zmq.NOBLOCK)

while True:
    while True:
        try:
            string = sub_socket.recv_string(flags=zmq.NOBLOCK)
            event_type = string[:7]
            source = string[7:4]
            val = string[11:]
            if event_type == "..CLAIM":
                claim(val)
            elif event_type == "RETRACT":
                retract(val)
        except zmq.Again:
            break
    time.sleep(0.1)

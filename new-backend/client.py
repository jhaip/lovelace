from parser import parse
from server import claim_fact, retract_fact, select_facts, get_facts_for_subscription
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

def claim(fact_string, source):
    logging.info("fact_string")
    logging.info(fact_string)
    logging.info("source")
    logging.info(source)
    fact = parse(fact_string, debug=True)
    logging.info("FACT:")
    logging.info(fact)
    claim_fact(fact, source)
    update_all_subscriptions()

def retract(fact_string):
    fact = parse(fact_string, debug=True)
    retract_fact(fact)
    update_all_subscriptions()

def send_results(source, id, results):
    results_str = json.dumps(results)
    pub_socket.send_string("{}{}{}".format(source, id, results_str), zmq.NOBLOCK)

def select(query_strings, select_id, source):
    query = map(lambda x: parse(x, debug=True), query_strings)
    facts = select_facts(query)
    send_results(source, select_id, facts)

def update_all_subscriptions():
    # Get all subscriptions
    query = [[('variable','source'),('text','subscription'),('variable','subscription_id'),('postfix','')]]
    subscriptions = select_facts(query)
    for row in subscriptions:
        source = row[0]
        subscription_id = row[1]
        facts = get_facts_for_subscription(source, subscription_id)
        send_results(source, subscription_id, facts)

def subscribe(fact_strings, subscription_id, source):
    for i, fact_string in enumerate(fact_strings):
        claim("subscription {} {} {}".format(subscription_id, i, fact_string), source)

sub_socket.setsockopt_string(zmq.SUBSCRIBE, "....CLAIM")
sub_socket.setsockopt_string(zmq.SUBSCRIBE, "...SELECT")
sub_socket.setsockopt_string(zmq.SUBSCRIBE, "..RETRACT")
sub_socket.setsockopt_string(zmq.SUBSCRIBE, "SUBSCRIBE")

while True:
    while True:
        try:
            string = sub_socket.recv_string(flags=zmq.NOBLOCK)
            logging.info("RECV: {}".format(string))
            event_type_len = 9
            source_len = 4
            event_type = string[:event_type_len]
            source = string[event_type_len:source_len]
            val = string[(event_type_len + source_len):]
            if event_type ==   "....CLAIM":
                claim(val, source)
            elif event_type == "..RETRACT":
                retract(val)
            elif event_type == "...SELECT":
                json_val = json.loads(val)
                select(json_val["facts"], json_val["id"], source)
            elif event_type == "SUBSCRIBE":
                json_val = json.loads(val)
                subscribe(json_val["facts"], json_val["id"], source)
        except zmq.Again:
            break
    logging.info("loop")
    time.sleep(0.5)

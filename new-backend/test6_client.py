import time
import zmq
import logging
import sys
import random

logging.basicConfig(level=logging.INFO)

context = zmq.Context()
rpc_url = "localhost"
sub_socket = context.socket(zmq.SUB)
sub_socket.connect("tcp://{0}:5556".format(rpc_url))
pub_socket = context.socket(zmq.PUB)
pub_socket.connect("tcp://{0}:5555".format(rpc_url))

MY_ID = sys.argv[1]

sub_socket.setsockopt_string(zmq.SUBSCRIBE, "SUBSCRIBE{}".format(MY_ID))

N = 10
i = N

time.sleep(1+random.random() * 0.1)

start = time.time()
pub_socket.send_string("....CLAIM{}".format(MY_ID), zmq.NOBLOCK)

while True:
    while True:
        try:
            string = sub_socket.recv_string(flags=zmq.NOBLOCK)
            # logging.info("RECV: {}".format(string))
            pub_socket.send_string("....CLAIM{}".format(MY_ID), zmq.NOBLOCK)
            i -= 1
            if i == 0:
                end = time.time()
                print("TIME: {} ms".format((end - start)*1000.0))
                sys.exit()
        except zmq.Again:
            break
    # logging.info("loop")
    # time.sleep(0.001)
    # time.sleep(0.5)

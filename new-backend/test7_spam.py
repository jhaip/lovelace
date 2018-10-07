import time
import zmq
import logging
import sys

logging.basicConfig(level=logging.INFO)

context = zmq.Context()
rpc_url = "localhost"
sub_socket = context.socket(zmq.SUB)
sub_socket.connect("tcp://{0}:5556".format(rpc_url))
pub_socket = context.socket(zmq.PUB)
pub_socket.connect("tcp://{0}:5555".format(rpc_url))

sub_socket.setsockopt_string(zmq.SUBSCRIBE, "SUBSCRIBE")

K = 1000

def send():
    for i in range(K):
        pub_socket.send_string("....CLAIM6BAD", zmq.NOBLOCK)

time.sleep(1)

start = time.time()

while True:
    send()
    time.sleep(0.1)
    logging.info("loop")
    # time.sleep(0.001)
    # time.sleep(0.5)

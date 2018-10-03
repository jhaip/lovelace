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

sub_socket.setsockopt_string(zmq.SUBSCRIBE, "....CLAIM")

MIN_ID = int(sys.argv[1])
MAX_ID = int(sys.argv[2])

while True:
    while True:
        try:
            string = sub_socket.recv_string(flags=zmq.NOBLOCK)
            string_id = int(string[9:])
            if string_id >= MIN_ID and string_id <= MAX_ID:
                # TODO: look at string and only send_string() to certain strings
                # logging.info("RECV: {}".format(string))
                pub_socket.send_string("SUBSCRIBE{}".format(string_id), zmq.NOBLOCK)
        except zmq.Again:
            break
    # logging.info("loop")
    # time.sleep(0.001)
    # time.sleep(0.5)

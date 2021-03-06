import time
import zmq
import logging

logging.basicConfig(level=logging.INFO)

context = zmq.Context()
rpc_url = "localhost"
sub_socket = context.socket(zmq.SUB)
sub_socket.connect("tcp://{0}:5556".format(rpc_url))
pub_socket = context.socket(zmq.PUB)
pub_socket.connect("tcp://{0}:5555".format(rpc_url))

sub_socket.setsockopt_string(zmq.SUBSCRIBE, "....CLAIM")

while True:
    while True:
        try:
            string = sub_socket.recv_string(flags=zmq.NOBLOCK)
            # logging.info("RECV: {}".format(string))
            pub_socket.send_string("SUBSCRIBE", zmq.NOBLOCK)
        except zmq.Again:
            break
    # logging.info("loop")
    # time.sleep(0.001)
    # time.sleep(0.5)

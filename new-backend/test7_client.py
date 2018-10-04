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

N = 100
i = N
K = 100

def send():
    pub_socket.send_string("....CLAIM5TEST", zmq.NOBLOCK)
    for i in range(K):
        pub_socket.send_string("....CLAIM6BAD", zmq.NOBLOCK)

time.sleep(1)

start = time.time()
send()

while True:
    while True:
        try:
            string = sub_socket.recv_string(flags=zmq.NOBLOCK)
            # logging.info("RECV: {}".format(string))
            send()
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

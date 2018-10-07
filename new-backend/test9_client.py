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

time.sleep(1)

start = time.time()
for i in range(10000):
    pub_socket.send_string("....CLAIM", zmq.NOBLOCK)

end = time.time()
print("TIME: {} ms".format((end - start)*1000.0))

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

logging.info("SLEEPING")
time.sleep(5.0)
logging.info("ALIVE!")

N = 0
start = time.time()

while True:
    string = sub_socket.recv_string()
    # logging.info("GOT ONE {}".format(N))
    N += 1
    a = string[1:]
    if N >= 10000:
        break

# while True:
#     while True:
#         try:
#             string = sub_socket.recv_string(zmq.NOBLOCK)
#             # logging.info("GOT ONE {}".format(N))
#             N += 1
#             if N >= 40000:
#                 break
#         except zmq.Again:
#             break
#     # logging.info("loop")
end = time.time()
print("TIME: {} ms".format((end - start)*1000.0))

import zmq
import random
import sys
import time

port = "5556"
context = zmq.Context()
socket = context.socket(zmq.PAIR)
socket.bind("tcp://*:%s" % port)

time.sleep(1)

while True:
    start_time = time.time()
    socket.send(b"Server message to client3", zmq.NOBLOCK)
    elapsed_time = time.time() - start_time
    print("send", elapsed_time*1000.0*1000.0, "ms")
    msg = socket.recv()
    print(msg)
    time.sleep(1)
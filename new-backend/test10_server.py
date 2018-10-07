import zmq
import random
import sys
import time

port = "5559"
context = zmq.Context()

socket = context.socket(zmq.PUB)
socket.bind("tcp://*:%s" % port)

# socket = context.socket(zmq.PAIR)
# socket.bind("tcp://*:%s" % port)

while True:
    start = time.time()
    socket.send(b"Server message to client3")
    end = time.time()
    print("TIME: {} ms".format((end - start)*1000.0))
    # msg = socket.recv()
    # print(msg)
    time.sleep(1)

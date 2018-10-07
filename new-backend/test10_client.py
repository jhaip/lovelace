import zmq
import random
import sys
import time

port = "5559"
context = zmq.Context()

# socket = context.socket(zmq.PAIR)
# socket.connect("tcp://localhost:%s" % port)

socket = context.socket(zmq.SUB)
socket.connect("tcp://localhost:%s" % port)
print("COnnected")
socket.setsockopt(zmq.SUBSCRIBE, b"Server")

while True:
    msg = socket.recv_string()
    print("loop")
    print(msg)
    # start = time.time()
    # socket.send(b"client message to server1")
    # end = time.time()
    # print("TIME: {} ms".format((end - start)*1000.0))
    # socket.send(b"client message to server2")
    time.sleep(1)

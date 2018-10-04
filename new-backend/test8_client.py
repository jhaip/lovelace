import os
from nanomsg import (
    PAIR,
    Socket
)
import time

SOCKET_ADDRESS = "tcp://127.0.0.1:5558"

N = 0

s1 = Socket(PAIR)
s1.connect(SOCKET_ADDRESS)
time.sleep(1.0)

start = time.time()
while True:
    # print("loop")
    s1.send(b'ABC')
    # print("sent")
    recieved = s1.recv()
    # print(recieved)
    N += 1
    # print(N)
    if N >= 100:
        end = time.time()
        print("TIME: {} ms".format((end - start)*1000.0))
        break
s1.close()

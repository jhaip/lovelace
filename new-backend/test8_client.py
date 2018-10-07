import os
from nanomsg import (
    PAIR,
    Socket
)
import time

SOCKET_ADDRESS = "inproc://test"

N = 0

s1 = Socket(PAIR)
s1.connect(SOCKET_ADDRESS)
print("BINDED")
time.sleep(1.0)

start = time.time()
while True:
    print("loop")
    s1.send(b'ABC')
    print("post")
    mid = time.time()
    # print("sent")
    recieved = s1.recv()
    # print(recieved)
    N += 1
    # print(N)
    if N >= 1:
        end = time.time()
        print("MID: {} ms".format((mid - start)*1000.0))
        print("TIME: {} ms".format((end - start)*1000.0))
        break
s1.close()

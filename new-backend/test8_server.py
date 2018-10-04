import os
from nanomsg import (
    PAIR,
    Socket
)

SOCKET_ADDRESS = "tcp://127.0.0.1:5558"

# with Socket(PAIR) as s1:
#     with Socket(PAIR) as s2:
#         s1.bind(SOCKET_ADDRESS)
#         s2.connect(SOCKET_ADDRESS)
#
#         sent = b'ABC'
#         s2.send(sent)
#         recieved = s1.recv()
#         print("DONE:")
#         print(recieved)
# self.assertEqual(sent, recieved)

s1 = Socket(PAIR)
s1.bind(SOCKET_ADDRESS)
while True:
    # print("loop")
    recieved = s1.recv()
    # print(recieved)
    s1.send(b'ABC')
    # print("sent")
s1.close()

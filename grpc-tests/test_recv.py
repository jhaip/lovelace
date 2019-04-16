import zmq
import time
import cv2
import numpy as np
import room_pb2

addr = 'tcp://*:5555'
ctx = zmq.Context()
s1 = ctx.socket(zmq.PAIR)

s1.bind(addr)

time.sleep(1.0)
print("waiting for data")

data = s1.recv()
print("got data")

room_update = room_pb2.RoomUpdate()
room_update.ParseFromString(data)

print(room_update.type)
print(room_update.tokens[0].type)
print(room_update.tokens[0].stringVal)
print(room_update.tokens[1].type)
print(room_update.tokens[1].stringVal)
print(room_update.tokens[2].type)
# print(room_update)

image_data = room_update.tokens[-1].bytesVal

print(len(image_data))

nparr = np.fromstring(image_data, np.uint8)
image = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
cv2.imwrite("1.jpg", image)
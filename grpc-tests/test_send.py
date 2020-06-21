import zmq
import time
import cv2
import room_pb2

camera = cv2.VideoCapture(0)
addr = 'tcp://192.168.1.342:5555'
ctx = zmq.Context()
s2 = ctx.socket(zmq.PAIR)
s2.connect(addr)
# time.sleep(1.0)
return_value, image = camera.read()
print("got image")
_, img_encoded = cv2.imencode('.jpg', image)

room_update = room_pb2.RoomUpdate()
room_update.type = "CLAIM"
token1 = room_update.tokens.add()
token1.type = "TEXT"
token1.stringVal = "frame"
token2 = room_update.tokens.add()
token2.type = "TEXT"
token2.stringVal = "is"
token3 = room_update.tokens.add()
token3.type = "BYTES"
token3.bytesVal = img_encoded.tostring()

s2.send(room_update.SerializeToString())

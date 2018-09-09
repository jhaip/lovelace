# 7/17/2018

POST localhost:3000/assert facts=camera 1 sees dots "[{'x': 0.15, 'y': 0.1, 'r': 255, 'g': 0, 'b': 0}, {'x': 0.2, 'y': 0.1, 'r': 0, 'g': 255, 'b': 0}]" @ 1

v4l2-ctl -d /dev/video0 --list-ctrls
v4l2-ctl -d /dev/video0 --set-ctrl=focus_auto=0

from pynput import keyboard
import requests
import time
import sys
import os
import logging
import math
import json

scriptName = os.path.basename(__file__)
scriptNameNoExtension = os.path.splitext(scriptName)[0]
fileDir = os.path.dirname(os.path.realpath(__file__))
logPath = os.path.join(fileDir, 'logs/' + scriptNameNoExtension + '.log')
print(logPath)

logging.basicConfig(filename=logPath, level=logging.INFO)

URL = "http://localhost:3000/"

def say(fact):
    payload = {'facts': fact}
    return requests.post(URL + "assert", data=payload)

def retract(fact):
    payload = {'facts': fact}
    return requests.post(URL + "retract", data=payload)

def select(fact):
    payload = {'facts': fact}
    response = requests.post(URL + "select", data=payload)
    return response.json()

###########

id = 999  # TODO
is_ctrl_pressed = False

retract("keyboard {} typed key $ @ $".format(id))
retract("keyboard {} typed special key $ @ $".format(id))

def map_special_key(key):
    m = {}
    m[keyboard.Key.backspace] = "backspace"
    m[keyboard.Key.enter] = "enter"
    m[keyboard.Key.tab] = "tab"
    m[keyboard.Key.space] = "space"
    m[keyboard.Key.left] = "left"
    m[keyboard.Key.right] = "right"
    m[keyboard.Key.up] = "up"
    m[keyboard.Key.down] = "down"
    m["C-p"] = "C-p"
    m["C-s"] = "C-s"
    if key in m:
        return m[key]
    return None


def add_key(key, special_key):
    timestamp = int(time.time()*1000.0)
    retract("keyboard {} typed key $ @ $".format(id))
    retract("keyboard {} typed special key $ @ $".format(id))
    if special_key:
        special_key = map_special_key(special_key)
        assert("keyboard {} typed key \"{}\" @ {}".format(id, key, timestamp))
    else:
        assert("keyboard {} typed special key \"{}\" @ {}".format(id, key, timestamp))

def add_ctrl_key_combo(key):
    add_key(None, "C-{0}".format(key))


def on_press(key):
    global is_ctrl_pressed
    try:
        logging.info('alphanumeric key {0} pressed'.format(
            key.char))
        if is_ctrl_pressed:
            add_ctrl_key_combo(key.char)
        else:
            add_key(key.char, None)
    except AttributeError:
        logging.info('special key {0} pressed'.format(
            key))
        add_key(None, key)
        if key == keyboard.Key.ctrl:
            is_ctrl_pressed = True

def on_release(key):
    global is_ctrl_pressed
    logging.info('{0} released'.format(
        key))
    if key == keyboard.Key.ctrl:
        is_ctrl_pressed = False
    if key == keyboard.Key.esc:
        # Stop listener
        return False

# Collect events until released
with keyboard.Listener(
        on_press=on_press,
        on_release=on_release) as listener:
    listener.join()

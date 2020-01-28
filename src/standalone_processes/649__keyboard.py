from helper2 import init, claim, retract, prehook, subscription, batch, get_my_id_str, check_server_connection
import keyboard
import time
import logging

init(__file__, skipListening=True)
batch([{"type": "retract", "fact": [["id", get_my_id_str()], ["postfix", ""]]}])

def add_key(key, special_key):
    logging.info("{} - {}".format(key, special_key))
    timestamp = int(time.time()*1000.0)
    claims = []
    claims.append({"type": "retract", "fact": [
        ["id", get_my_id_str()],
        ["postfix", ""],
    ]})
    if special_key:
        logging.info("ADDING SPECIAL KEY {}".format(special_key))
        claims.append({"type": "claim", "fact": [
            ["id", get_my_id_str()],
            ["id", "0"],
            ["text", "keyboard"],
            ["text", get_my_id_str()],
            ["text", "typed"],
            ["text", "special"],
            ["text", "key"],
            ["text", str(special_key)],
            ["text", "@"],
            ["integer", str(timestamp)],
        ]})
    else:
        logging.info("ADDING KEY {}".format(key))
        claims.append({"type": "claim", "fact": [
            ["id", get_my_id_str()],
            ["id", "0"],
            ["text", "keyboard"],
            ["text", get_my_id_str()],
            ["text", "typed"],
            ["text", "key"],
            ["text", str(key)],
            ["text", "@"],
            ["integer", str(timestamp)],
        ]})
    batch(claims)

def handle_key_event(e):
    ctrl_held = keyboard.is_pressed('ctrl')
    shift_held = keyboard.is_pressed('shift')
    if e.event_type == 'down':
        if e.name == 'unknown':
            return
        if ctrl_held:
            add_key(None, 'C-{}'.format(e.name))
        elif shift_held and e.name.isalpha():
            add_key(e.name.upper(), None)
        else:
            special_keys = ['backspace', 'enter', 'tab', 'space', 'left', 'right', 'up', 'down', 'shift']
            if e.name in special_keys:
                add_key(None, e.name)
            else:
                add_key(e.name, None)

keyboard.hook(handle_key_event)
keyboard.wait()


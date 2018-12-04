import time
import logging
import pygame
from helper2 import init, claim, retract, prehook, subscription, batch, get_my_id_str
init(__file__, skipListening=True)
batch([{"type": "retract", "fact": [["id", get_my_id_str()], ["postfix", ""]]}])

def add_key(key, special_key):
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
            ["text", "keyboard"],
            ["text", get_my_id_str()],
            ["text", "typed"],
            ["text", "key"],
            ["text", str(key)],
            ["text", "@"],
            ["integer", str(timestamp)],
        ]})
    batch(claims)

pygame.init()
screen = pygame.display.set_mode((50, 50), 0, 32)
while True:
    pressed = pygame.key.get_pressed()
    ctrl_held = pressed[pygame.K_LCTRL] or pressed[pygame.K_RCTRL]
    for event in pygame.event.get():
        if event.type == pygame.KEYDOWN:
            if event.key == pygame.K_BACKSPACE:
                add_key(None, "backspace")
            elif event.key == pygame.K_RETURN:
                add_key(None, "enter")
            elif event.key == pygame.K_TAB:
                add_key(None, "tab")
            elif event.key == pygame.K_SPACE:
                add_key(None, "space")
            elif event.key == pygame.K_LEFT:
                add_key(None, "left")
            elif event.key == pygame.K_RIGHT:
                add_key(None, "right")
            elif event.key == pygame.K_UP:
                add_key(None, "up")
            elif event.key == pygame.K_DOWN:
                add_key(None, "down")
            elif event.key == pygame.K_n and ctrl_held:
                add_key(None, "C-n")
            elif event.key == pygame.K_p and ctrl_held:
                add_key(None, "C-p")
            elif event.key == pygame.K_s and ctrl_held:
                add_key(None, "C-s")
            elif event.key == pygame.K_EQUALS and ctrl_held:
                add_key(None, "C-+")
            elif event.key == pygame.K_MINUS and ctrl_held:
                add_key(None, "C--")
            elif event.key == pygame.K_QUOTEDBL:
                add_key(None, "\"")
            elif len(event.unicode) > 0:
                add_key(event.unicode, None)

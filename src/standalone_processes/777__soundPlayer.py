from helper2 import init, claim, retract, prehook, subscription, batch, MY_ID_STR, listen
import pygame
import pygame.midi
import logging
import time

BPM = 60
MELODY = []
MELODY_INDEX = 0
INSTRUMENT = 0

@subscription(["$ melody is %melody"])
def melody_callback(results):
    global MELODY, MELODY_INDEX
    logging.error("results:")
    logging.error(results)
    if len(results) == 0:
        MELODY = []
        return
    result = results[0]
    if not result or result.get("melody", '') == '':
        MELODY = []
        return
    try:
        result_melody = result["melody"].replace(" ", "").split(",")
        result_melody = list(map(lambda x: int(x), result_melody))
        MELODY_INDEX = 0
        MELODY = result_melody
        logging.error("new melody:")
        logging.error(MELODY)
    except:
        logging.error("bad melody {}".format(result["melody"]))

@subscription(["$ beats per minute is $bpm"])
def bpm_callback(results):
    global BPM
    logging.error("results:")
    logging.error(results)
    if len(results) == 0:
        BPM = 60
        return
    result = results[0]
    try:
        result_bpm = int(result["bpm"])
        BPM = result_bpm
        logging.error("new bpm:")
        logging.error(BPM)
    except:
        logging.error("bad BPM {}".format(result["bpm"]))


@subscription(["$ instrument is $instrument"])
def instrument_callback(results):
    global INSTRUMENT
    logging.error("results:")
    logging.error(results)
    if len(results) == 0:
        INSTRUMENT = 0
        return
    try:
        result_instrument = int(results[0]["instrument"])
        if result_instrument > 127 or result_instrument < 0:
            result_instrument = 0
        INSTRUMENT = result_instrument
        logging.error("new instrument:")
        logging.error(INSTRUMENT)
    except:
        logging.error("BAD instrument {}".format(results))

def playNextNode(midi_out, instrument, pitch, duration):
    logging.info("playing stuff")
    midi_out.set_instrument(instrument)
    # 74 is middle C, 127 is "how loud" - max is 127
    midi_out.note_on(pitch, 127)
    time.sleep(duration)
    midi_out.note_off(pitch, 127)
    # time.sleep(.5)


init(__file__, skipListening=True)

pygame.init()
pygame.midi.init()

port = pygame.midi.get_default_output_id()
logging.info("using output_id :%s:" % port)
midi_out = pygame.midi.Output(port, 0)

try:
    lastUpdateTime = time.time()
    while True:
        listen()
        if time.time() - lastUpdateTime >= 60.0/BPM and len(MELODY) > 0:
            playNextNode(midi_out, INSTRUMENT, MELODY[MELODY_INDEX], 60.0/BPM)
            MELODY_INDEX = (MELODY_INDEX + 1) % len(MELODY)
        else:
            time.sleep(0.01)
finally:
    midi_out.note_off(MELODY[MELODY_INDEX], 127)
    midi_out.close()
    del midi_out
    pygame.midi.quit()

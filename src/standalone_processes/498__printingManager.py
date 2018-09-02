import subprocess
import requests
import time
import sys
import os
import logging

scriptName = os.path.basename(__file__)
scriptNameNoExtension = os.path.splitext(scriptName)[0]
fileDir = os.path.dirname(os.path.realpath(__file__))
logPath = os.path.join(fileDir, 'logs/' + scriptNameNoExtension + '.log')
print(logPath)

logging.basicConfig(filename=logPath, level=logging.INFO)

logging.info("printingManager started!")

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

while True:
    logging.info("checking for printing wishes")
    print_wishes = select('wish file $name would be printed')
    for wish in print_wishes:
        name = wish['name']['word']
        retract('wish file {} would be printed'.format(name))
        logging.info("PRINTING:", name)
        subprocess.call(['/usr/bin/lpr', name])
    time.sleep(1)

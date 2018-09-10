import requests
import time
import sys
import os
import logging
import math
import json

scriptName = None
URL = "http://localhost:3000/"
MY_ID = None

def init(root_filename):
    global MY_ID
    scriptName = os.path.basename(root_filename)
    scriptNameNoExtension = os.path.splitext(scriptName)[0]
    fileDir = os.path.dirname(os.path.realpath(root_filename))
    logPath = os.path.join(fileDir, 'logs/' + scriptNameNoExtension + '.log')
    logging.basicConfig(filename=logPath, level=logging.INFO)
    MY_ID = (scriptName.split(".")[0]).split("__")[0]

def say(fact, targetPaper=None):
    global MY_ID
    if targetPaper is None:
        targetPaper = MY_ID
    payload = {'facts': targetPaper + ' ' + fact}
    return requests.post(URL + "assert", data=payload)

def retract(fact, targetPaper=MY_ID):
    global MY_ID
    if targetPaper is None:
        targetPaper = MY_ID
    payload = {'facts': targetPaper + ' ' + fact}
    return requests.post(URL + "retract", data=payload)

def select(fact, targetPaper='$'):
    payload = {'facts': targetPaper + ' ' + fact}
    response = requests.post(URL + "select", data=payload)
    return response.json()

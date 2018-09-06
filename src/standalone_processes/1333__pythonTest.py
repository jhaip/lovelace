import requests
import time

URL = 'http://localhost:3000/'

def say(fact):
    payload = {'facts': fact}
    return requests.post(URL + 'assert', data=payload)

def retract(fact):
    payload = {'facts': fact}
    return requests.post(URL + 'retract', data=payload)

def select(fact):
    payload = {'facts': fact}
    response = requests.post(URL + 'select', data=payload)
    return response.json()


print('Hello from pythonTest.py')
retract('hello from testProcess @ $')

programs = select('$program is active')
active_programs = []
for program in programs:
    active_programs.append(program['program']['word'])
print('Active programs:')
print(active_programs)

while True:
    print('hello from pythonTest.p', time.time())
    retract('hello from pythonTest.py @ $')
    say('hello from pythonTest.py @ {}'.format(int(round(time.time() * 1000))))
    time.sleep(1)

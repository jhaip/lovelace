from helper import *
init(__file__)

logging.info('Hello from pythonTest.py')
retract('hello from testProcess @ $')

programs = select('$program has paper ID $paperId')
active_programs = []
for program in programs:
    active_programs.append(program['program']['word'])
logging.info('programs:')
logging.info(active_programs)

while True:
    logging.info('hello from pythonTest.')
    logging.info(time.time())
    retract('hello from pythonTest.py @ $')
    say('hello from pythonTest.py @ {}'.format(int(round(time.time() * 1000))))
    time.sleep(1)

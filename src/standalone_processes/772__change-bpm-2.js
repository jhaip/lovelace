const { room, myId, scriptName, run } = require('../helper2')(__filename)

room.cleanup()
room.retract("$ beats per minute is $")
room.assert("beats per minute is 500")

run()
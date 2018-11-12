const { room, myId, scriptName, run } = require('../helper2')(__filename)

room.cleanup()
room.retract("$ instrument is $")
room.assert("instrument is 19")

run()
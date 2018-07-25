const Room = require('@living-room/client-js')
const room = new Room()

console.log("start testProcess")
room.retract(`hello from testProcess @ $`)

setInterval(() => {
  console.error("hello from testProcess", new Date())
  room
    .retract(`hello from testProcess @ $`)
    .assert(`hello from testProcess @ ${(new Date()).getTime()}`)
}, 1000)

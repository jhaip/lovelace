const { room, myId } = require('../helper')(__filename);

console.log("start taco")
room.retract(`#${myId} hello from taco @ $`)

setInterval(() => {
  console.error("hello from taco", new Date())
  room
    .retract(`#${myId} hello from taco @ $`)
    .assert(`#${myId} hello from taco @ ${(new Date()).getTime()}`)
}, 1000);

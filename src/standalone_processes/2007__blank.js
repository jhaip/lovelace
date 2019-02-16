const { room, myId, run } = require('../helper2')(__filename);

room.on("$ it is go time", results => {
  room.cleanup()
  if(results.length > 0) {
    room.assert("wish I was labeled IT IS GO TIME")
  } else {
    room.assert("wish I was labeled waiting for go time")
  }
})

run()
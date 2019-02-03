const { room, myId, run } = require('../helper2')(__filename);

room.on(`time is $time`, results => {
  room.cleanup()
  let current_time = results.length > 0 ? results[0].time : 1;
  if (Math.floor(current_time/1000) % 2 === 0) {
    room.assert("I wish I was highlighted orange")
  }
})

run()
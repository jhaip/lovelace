const { room, myId, run } = require('../helper2')(__filename)

room.on(`$ time is $time`, results => {
  room.cleanup();
  if (results.length === 0) return
  let current_time = new Date(results[0].time);
  room.assert(`wish I was labeled ${current_time}`)
})

run();
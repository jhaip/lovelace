const { room, myId, run } = require('../helper2')(__filename);

room.onRaw(`$source I wish I was highlighted $color`, results => {
  room.cleanup();
  results.forEach(result => {
    let ill = room.newIllumination()  
    ill.fill(result.color)
    ill.rect(0, 0, 1000, 1000);
    room.draw(ill, result.source)
  })
})

run()
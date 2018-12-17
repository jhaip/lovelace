const { room, myId, run } = require('../helper2')(__filename);

// Write code here!

room.cleanup();

let ill = room.newIllumination()
ill.fill("red")
ill.rect(0, 0, 1000, 1000);
room.draw(ill)

run()
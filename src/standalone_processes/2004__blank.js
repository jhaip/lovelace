const { room, myId, run } = require('../helper2')(__filename);

// Write code here!
const DELAY_MS = 100;

function tick() {
  room.cleanup()
  room.assert(`time is ${(new Date()).getTime()}`);
  let ill = room.newIllumination()
  ill.text(0, 100, "I am a clock")
  room.draw(ill);
  room.flush();
  setTimeout(tick, DELAY_MS);
}

tick();

run();
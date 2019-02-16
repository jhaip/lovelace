const { room, myId, run } = require('../helper2')(__filename);

const DELAY_MS = 500;

function tick() {
  room.cleanup()
  room.assert(`time is ${(new Date()).getTime()}`);
  room.assert(`wish I was labeled This clock ticks once every ${DELAY_MS}ms`)
  room.flush();
  setTimeout(tick, DELAY_MS);
}

tick();

run();

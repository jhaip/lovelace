const { room, myId, run } = require('../helper2')(__filename);

board = []
W = 27
H = 27

const illb = room.newIllumination()
illb.push()
illb.rotate(Math.PI/2.0)
illb.translate(0, -H*4)
illb.nofill()
illb.rect(0, 0, W*6, H*4)
illb.pop()
room.draw(illb)
const tile_color = {
  "up": [255, 0, 0],
  "down": [0, 255, 0],
  "right": [255, 100, 0],
  "left": [255, 0, 100],
  "loopstop": [0, 128, 255],
  "loopstart": [128, 0, 255],
}

room.on(`tile $tile seen at $x $y @ $t`,
        results => {
  room.subscriptionPrefix(1);
  if (!!results) {
    results.forEach(({ tile, x, y, t }) => {
    room.assert(`wish`)
    let ill = room.newIllumination()
    ill.push()
    ill.rotate(Math.PI/2.0)
    ill.translate(0, -H*4)
    if (tile in tile_color) {
        c = tile_color[tile]
        ill.fill(c[0], c[1], c[2])
    } else {
        ill.nofill();
    }
    ill.rect(x*W, y*H, W, H)    
    ill.pop()
    room.draw(ill)


    });
  }
  room.subscriptionPostfix();
})


run();

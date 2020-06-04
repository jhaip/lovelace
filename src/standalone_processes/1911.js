const { room, myId, run } = require('../helper2')(__filename);

var C = [1, 0, 0, 0, 1, 0, 0, 0, 1];

room.on(`camera $cam calibration for $display is $M1 $M2 $M3 $M4 $M5 $M6 $M7 $M8 $M9`,
        results => {
  room.subscriptionPrefix(1);
  if (!!results && results.length > 0) {
    results.forEach(({ cam, display, M1, M2, M3, M4, M5, M6, M7, M8, M9 }) => {
C = [+M1, +M2, +M3, +M4, +M5, +M6, +M7, +M8, +M9]

    });
  }
  room.subscriptionPostfix();
})

room.on(`laser seen at $x $y @ $t on camera $cam`,
        results => {
  room.subscriptionPrefix(2);
  if (!!results && results.length > 0) {
    results.forEach(({ x, y, t, cam }) => {

let ill = room.newIllumination()
ill.set_transform(C[0], C[1], C[2], C[3], C[4], C[5], C[6], C[7], C[8])
ill.fill(255, 255, 255)
ill.stroke(255, 0, 0);
ill.ellipse(+x-20, +y-20, 40, 40)
room.draw(ill, "1997")


    });
  }
  room.subscriptionPostfix();
})


run();

const { room, myId, run } = require('../helper2')(__filename);

room.cleanup()
room.assert(`new it is working again`)

room.on(`$photon sees color ( $r, $g, $b )`,
        results => {
  subscriptionPrefix();
  if (!!results) {
    results.forEach(({ photon, r, g, b }) => {
    room.assert(`wish tablet had background color ( ${r} , ${g} , ${b} )`)

    });
  } else {
    room.assert(`wish tablet had background color ( 0 , 0 , 0 )`)
  }
  subscriptionPostfix();
})

let x = 5;

room.on(`x is ${x}`,
        `ok`,
        results => {
  subscriptionPrefix();
  if (!!results) {
    results.forEach(({  }) => {
  // yo

    });
  } else {
  // ok
  }
  subscriptionPostfix();
})


run();

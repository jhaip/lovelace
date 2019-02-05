const { room, myId, run } = require('../helper2')(__filename);

room.cleanup()
room.assert(`it is working`)

room.on(`$photon sees color ( $r, $g, $b )`,
        results => {
subscriptionPrefix();
if (!!results) {

    // I'm new
    // room.cleanup()
    room.assert(`wish tablet had background color ( 0 , 0 , 0 )`)
} else {
    // room.cleanup()
    let r = results[0].r;
    let g = results[0].g;
    let b = results[0].b;
    room.assert(`wish tablet had background color ( ${r} , ${g} , ${b} )`)
}
subscriptionPostfix();
})

let x = 5;

room.on(`x is ${x}`,
        `ok`,
        results => {
subscriptionPrefix();
if (!!results) {

  // yo
} else {
  // ok

}
subscriptionPostfix();
})


run();
const { room, myId, run } = require('../helper2')(__filename);

room.cleanup()
room.assert(`it is working`)

// new

room.on(`$photon sees color ( $r, $g, $b )`,
        results => {
if (!results) {

    // I'm new
    room.cleanup()
    room.assert(`wish tablet had background color ( 0 , 0 , 0 )`)
} else {
    room.cleanup()
    let r = results[0].r;
    let g = results[0].g;
    let b = results[0].b;
    room.assert(`wish tablet had background color ( ${r} , ${g} , ${b} )`)
}
})


run();
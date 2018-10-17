const { room, myId, scriptName } = require('../helper2')(__filename);

let N = 100
let i = 1

room.on(
    `$X Fox is out`,
    results => {
        console.log("results:")
        console.log(results)
        console.log("Fox is out!")
    }
)

setTimeout(() => {
    room.assert(`Fox is out`)
    setTimeout(() => {
        room.assert(`Fox is out`)
        setTimeout(() => {
            room.retract(`$X Fox is out`)
            setTimeout(() => {
                room.assert(`Fox is out`)
                setTimeout(() => {
                    room.assert(`Fox is out`)
                }, 1000);
            }, 1000);
        }, 1000);
    }, 1000);
}, 2000);
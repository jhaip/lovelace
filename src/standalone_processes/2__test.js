const { room, myId, scriptName, MY_ID_STR } = require('../helper2')(__filename);

let N = 100
let i = 1

room.on(
    `$ $X ${myId} has $Y toes`,
    results => {
        room.retract(`#${MY_ID_STR} %`)
        console.log("results:")
        console.log(results)
        i += 1;
        if (i >= N) {
            console.log("\n...DONE!")
            process.exit(0);
        }
        room.assert(`Man ${(myId * 1.0).toFixed(6)} has ${i} toes`)
        room.assert(`Man ${(myId * 1.0).toFixed(6)} has ${i} toes`)
        room.flush()
    }
)

setTimeout(() => {
    room.assert(`Man ${(myId * 1.0).toFixed(6)} has 0 toes`, true)
}, 2000);
const { room, myId, scriptName } = require('../helper2')(__filename);

let N = 100
let i = 1

room.on(
    `$ $X ${myId} has $Y toes`,
    results => {
        console.log("results:")
        console.log(results)
        i += 1;
        if (i >= N) {
            console.log("\n...DONE!")
            process.exit(0);
        }
        room.assert(`Man ${myId} has ${i} toes`)
        room.assert(`Man ${myId} has ${i} toes`)
    }
)

setTimeout(() => {
    room.assert(`Man ${myId} has 0 toes`)
}, 2000);
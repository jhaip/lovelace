const { room, myId, run, MY_ID_STR } = require('../helper2')(__filename);

const N = 10;
const F = parseInt(myId) + N;
console.log("testing with N =", N);

room.on(
    `test client ${myId} says $x @ $time1`,
    `test client ${F} says $y @ $time2`,
    results => {
        console.error(results);
        room.subscriptionPrefix(1);
        if (!!results) {
            const currentTimeMs = (new Date()).getTime()
            console.error(`TEST IS DONE @ ${currentTimeMs}`)
            console.log("elapsed time:", parseInt(results[0].time2) - parseInt(results[0].time1), "ms")
        }
        room.subscriptionPostfix();
    }
)

const currentTimeMs = (new Date()).getTime()
room.assert(`test client ${myId} says ${myId} @ ${currentTimeMs}`);

run()

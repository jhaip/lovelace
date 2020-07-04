const spawn = require('child_process').spawn;
const { room, myId, run, MY_ID_STR, getIdFromProcessName, getIdStringFromId } = require('../helper2')(__filename);

room.onGetSource('wisherId',
    `wish speaker said $text`,
    results => {
        room.subscriptionPrefix(1);
        if (!!results && results.length > 0) {
            results.forEach(({ wisherId, text }) => {
                runArgs = ["-ven+m7", text, "2>/dev/null"]
                console.log(runArgs)
                const child = spawn('espeak', runArgs)
                console.log("done")
                room.retractFromSource(wisherId, `wish speaker said $`)
            });
        }
        room.subscriptionPostfix();
    })

run();

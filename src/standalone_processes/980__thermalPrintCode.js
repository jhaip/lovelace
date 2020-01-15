const { room, myId, run } = require('../helper2')(__filename);

room.on(
    `wish paper $id at $shortFilename would be printed`,
    `$shortFilename has source code $sourceCode`,
    results => {
        room.subscriptionPrefix(1);
        if (!!results) {
            results.forEach(({ id, shortFilename, sourceCode }) => {
                room.assert(`wish text`, ["text", sourceCode], `would be thermal printed`)
            });
        }
        room.subscriptionPostfix();
    })


run();

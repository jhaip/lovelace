const express = require('express')
const bodyParser = require("body-parser");
const { room, myId, scriptName, run } = require('../helper2')(__filename);
const app = express();
const port = 5000;

app.use(bodyParser.urlencoded({ extended: false }));
app.use(bodyParser.json());

app.post('/cleanup-claim', (req, res) => {
    console.error("cleanup-claim")
    console.error(req.body)
    if (Array.isArray(req.body.retract)) {
        room.retract(...req.body.retract)
    } else {
        room.retract(req.body.retract)
    }
    if (Array.isArray(req.body.claim)) {
        room.assert(...req.body.claim)
    } else {
        room.assert(req.body.claim)
    }
    room.flush()
    res.status(204).send("");
})

app.listen(port, () => console.log(`Example app listening on port ${port}!`))
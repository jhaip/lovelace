const express = require('express')
const bodyParser = require("body-parser");
const { room, myId, scriptName, run } = require('../helper2')(__filename);
const app = express();
const port = 3014;

app.use(bodyParser.urlencoded({ extended: false }));
app.use(bodyParser.json());
app.use(express.static('./src/web-display'))

var graphicsCache = [];
var calibration = null;

app.get('/status', (req, res) => {
    res.status(200).send({
        'calibration': calibration,
        'graphics': graphicsCache
    });
})

room.on(`draw graphics $graphics on web2`,
    results => {
        room.subscriptionPrefix(2);
        if (!!results) {
            graphicsCache = [];
            results.forEach(({ graphics }) => {
                let parsedGraphics = JSON.parse(graphics)
                graphicsCache = graphicsCache.concat(parsedGraphics);
            });
        }
        room.subscriptionPostfix();
    })

app.listen(port, () => console.log(`Example app listening on port ${port}!`));
run();

const express = require('express')
const bodyParser = require("body-parser");
const fs = require('fs');
const { room, myId, scriptName, run } = require('../helper2')(__filename);
const app = express();
const port = 3000;

app.use(bodyParser.urlencoded({ extended: false }));
app.use(bodyParser.json());

app.get('/db', (req, res) => {
    // res.send('Hello World!')
    fs.readFile('./new-backend/go-server/db_view_base64.txt', 'utf8', function (err, contents) {
        console.log(contents);
        const l = contents.split("\n");
        console.log(l);
        res.send(l);
    });
})

app.post('/select', (req, res) => {
    const query_strings = req.body.query;
    if (!query_strings) {
        res.status(400).send('Missing query')
    }
    console.log("query strings:");
    console.log(query_strings);
    room.on("$ camera $cameraId sees paper $id at TL ($x1, $y1) TR ($x2, $y2) BR ($x3, $y3) BL ($x4, $y4) @ $time",
        results => {
            console.log("RESULTS:")
            console.log(results);
            // cleanup also removes the subscription in the fact database
            room.cleanup();
            room.flush();
            console.log("sending results:")
            res.send(results);
        }
    )
    // if no results are returned within 5 seconds, return []
    // no results are returned if there are no matched facts at this time
    setTimeout(() => {
        room.cleanup();
        room.flush();
        res.status(200).send([])
    }, 5000);
})

app.listen(port, () => console.log(`Example app listening on port ${port}!`))
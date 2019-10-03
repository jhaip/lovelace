const express = require('express')
const bodyParser = require("body-parser");
const fs = require('fs');
const { room, myId, scriptName, run } = require('../helper2')(__filename);
const app = express();
const port = 3011;

app.use(bodyParser.urlencoded({ extended: false }));
app.use(bodyParser.json());
app.use(express.static('./src/region-editor'))

var regionData = [{
    'id': '9df78dc0-9e97-4a63-851e-b5bd61ba55c6',
    'name': 'pl1health',
    'x1': 20 * 6,
    'y1': 20 * 6,
    'x2': 100 * 6,
    'y2': 20 * 6,
    'x3': 100 * 6,
    'y3': 100 * 6,
    'x4': 20 * 6,
    'y4': 100 * 6,
    'toggleable': true
}];

app.get('/status', (req, res) => {
    res.status(200).send(regionData);
})

app.listen(port, () => console.log(`Example app listening on port ${port}!`))
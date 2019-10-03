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

app.delete('/region/:regionId', (req, res) => {
    const regionId = req.params.regionId;
    room.retractAll(`region "${regionId}" %`);
    room.flush();
    res.status(200);
})

app.put('/region/:regionId', (req, res) => {
    const regionId = req.params.regionId;
    const data = req.body.data;
    if (typeof data.name !== "undefined") {
        room.retractAll(`region "${regionId}" has name $`);
        room.assert(`region "${regionId}" has name "${data.name}"`);
    }
    if (typeof data.toggleable !== "undefined") {
        room.retractAll(`region "${regionId}" is toggleable`);
        if (data.toggleable) {
            room.assert(`region "${regionId}" is toggleable`);
        }
    }
    if (typeof data.x1 !== "undefined") {
        room.retractAll(`region "${regionId}" at %`);
        room.assert(`region "${regionId}" at ${data.x1} ${data.y1} ${data.x2} ${data.y2} ${data.x3} ${data.y3} ${data.x4} ${data.y4}`);
    }
    room.flush();
    res.status(200);
})

room.on(`region $id at $x1 $y1 $x2 $y2 $x3 $y3 $x4 $y4`,
    results => {
        room.subscriptionPrefix(2);
        if (!!results) {
            let seenRegions = {};
            results.forEach(result => {
                seenRegions[result.id] = true;
                let regionUpdated = false;
                for (let i = 0; i < regionData.length; i+=1) {
                    if (regionData[i].id === result.id) {
                        regionData[i] = Object.assign(regionData[i], result)
                        regionUpdated = true;
                        break;
                    }
                }
                if (!regionUpdated) {
                    regionData.push(result);
                }
            });
            regionData = regionData.filter(r => !!seenRegions[r.id]);
        }
        room.subscriptionPostfix();
    })

room.on(`region $id is toggleable`,
    results => {
        room.subscriptionPrefix(3);
        if (!!results) {
            results.forEach(result => {
                let regionUpdated = false;
                result.toggleable = true;
                for (let i = 0; i < regionData.length; i += 1) {
                    if (regionData[i].id === result.id) {
                        regionData[i] = Object.assign(regionData[i], result)
                        regionUpdated = true;
                        break;
                    }
                }
                if (!regionUpdated) {
                    regionData.push(result);
                }
            });
        }
        room.subscriptionPostfix();
    })

room.on(`region $id has name $name`,
    results => {
        room.subscriptionPrefix(4);
        if (!!results) {
            results.forEach(result => {
                let regionUpdated = false;
                for (let i = 0; i < regionData.length; i += 1) {
                    if (regionData[i].id === result.id) {
                        regionData[i] = Object.assign(regionData[i], result)
                        regionUpdated = true;
                        break;
                    }
                }
                if (!regionUpdated) {
                    regionData.push(result);
                }
            });
        }
        room.subscriptionPostfix();
    })

app.listen(port, () => console.log(`Example app listening on port ${port}!`));
run();
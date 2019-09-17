var flatbuffers = require('./flatbuffers').flatbuffers;
var roomupdatefbs = require('./request_generated').roomupdatefbs; // Generated by `flatc`.

function stringToUint(string) {
    var string = unescape(encodeURIComponent(string)),
        charList = string.split(''),
        uintArray = [];
    for (var i = 0; i < charList.length; i++) {
        uintArray.push(charList[i].charCodeAt(0));
    }
    return new Uint8Array(uintArray);
}

function uintToString(uintArray) {
    var encodedString = String.fromCharCode.apply(null, uintArray),
        decodedString = decodeURIComponent(escape(encodedString));
    return decodedString;
}

function testRoomUpdateSerialization() {
    var builder = new flatbuffers.Builder(1024);

    var factValue = roomupdatefbs.Fact.createValueVector(builder, stringToUint("Hello World!"))

    roomupdatefbs.Fact.startFact(builder)
    roomupdatefbs.Fact.addType(builder, roomupdatefbs.FactTypeText)
    roomupdatefbs.Fact.addValue(builder, factValue)
    var fact = roomupdatefbs.Fact.endFact(builder)

    var updateSource = builder.createString("1234")
    var updateSubId = builder.createString("asdf")
    var facts = roomupdatefbs.RoomUpdate.createFactsVector(builder, [fact])

    roomupdatefbs.RoomUpdate.startRoomUpdate(builder)
    roomupdatefbs.RoomUpdate.addType(builder, roomupdatefbs.UpdateTypeClaim)
    roomupdatefbs.RoomUpdate.addSource(builder, updateSource)
    roomupdatefbs.RoomUpdate.addSubscriptionId(builder, updateSubId)
    roomupdatefbs.RoomUpdate.addFacts(builder, facts)
    var update = roomupdatefbs.RoomUpdate.endRoomUpdate(builder)

    var updates = roomupdatefbs.RoomUpdates.createUpdatesVector(builder, [update])

    roomupdatefbs.RoomUpdates.startRoomUpdates(builder)
    roomupdatefbs.RoomUpdates.addUpdates(builder, updates)
    var full_updates_msg = roomupdatefbs.RoomUpdates.endRoomUpdates(builder)

    builder.finish(full_updates_msg)
    var msg_buf = builder.asUint8Array(); // Of type `Uint8Array`.
    return msg_buf
}

function testRoomUpdateDeserialization(data) {
    var buf = new flatbuffers.ByteBuffer(data);
    var room_updates_obj = roomupdatefbs.RoomUpdates.getRootAsRoomUpdates(buf)
    var updates_length = room_updates_obj.updatesLength()
    console.log(updates_length)
    update = room_updates_obj.updates(0)
    console.log(update.type())
    var source = update.source()
    var sub_id = update.subscriptionId()
    console.log(source)
    console.log(sub_id)
    facts_length = update.factsLength()
    console.log(facts_length)
    fact = update.facts(0)
    console.log(fact.type())
    console.log(fact.valueArray())
    console.log(uintToString(fact.valueArray()))
}

var s = testRoomUpdateSerialization()
console.log(s)
console.log(s.join(' '))
console.log(s.length)
console.log(testRoomUpdateDeserialization(s))

/*
JS:
12 0 0 0 0 0 6 0 8 0 4 0 6 0 0 0 4 0 0 0 1 0 0 0 16 0 0 0 12 0 20 0 19 0 12 0 8 0 4 0 12 0 0 0 16 0 0 0 20 0 0 0 28 0 0 0 0 0 0 1 1 0 0 0 36 0 0 0 4 0 0 0 97 115 100 102 0 0 0 0 4 0 0 0 49 50 51 52 0 0 0 0 8 0 12 0 11 0 4 0 8 0 0 0 8 0 0 0 0 0 0 1 12 0 0 0 72 101 108 108 111 32 87 111 114 108 100 33
Golang:
12 0 0 0 0 0 6 0 8 0 4 0 6 0 0 0 4 0 0 0 1 0 0 0 16 0 0 0 12 0 20 0 19 0 12 0 8 0 4 0 12 0 0 0 16 0 0 0 20 0 0 0 28 0 0 0 0 0 0 0 1 0 0 0 36 0 0 0 4 0 0 0 97 115 100 102 0 0 0 0 4 0 0 0 49 50 51 52 0 0 0 0 8 0 12 0 11 0 4 0 8 0 0 0 8 0 0 0 0 0 0 0 12 0 0 0 72 101 108 108 111 32 87 111 114 108 100 33
Difference:
--------------------------------------------------------------------------------------------------------------------------------1---------------------------------------------------------------------------------------------------------------------1-----------------------------------------------------
*/

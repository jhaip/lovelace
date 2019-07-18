const { room, myId, run } = require('../helper2')(__filename);

room.assert(`paper 2000 has RFID "f26a0c2e"`)
room.assert(`paper 2001 has RFID "f238222e"`)
room.assert(`paper 2002 has RFID "80616ea3"`)
room.assert(`paper 2003 has RFID "d07911a3"`)
room.assert(`paper 2004 has RFID "91b4d108"`)
room.assert(`paper 2005 has RFID "53825027"`)
room.assert(`paper 2006 has RFID "10af78a3"`)
room.assert(`paper 2007 has RFID "7341a727"`)
room.assert(`paper 2008 has RFID "2574c72d"`)
room.assert(`paper 2009 has RFID "b680cc21"`)
room.assert(`paper 2010 has RFID "737e9c27"`)
room.assert(`paper 2011 has RFID "4221cd24"`)
room.assert(`paper 5 has RFID "d01ff625"`)
room.assert(`paper 1100 has RFID "e21eef27"`)
room.assert(`paper 1013 has RFID "7bdbe359"`)

run();
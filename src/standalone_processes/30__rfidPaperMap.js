const { room, myId, run } = require('../helper2')(__filename);

room.assert(`paper 2000 has RFID`, ["text", "f26a0c2e"])
room.assert(`paper 2001 has RFID`, ["text", "f238222e"])
room.assert(`paper 2002 has RFID`, ["text", "80616ea3"])
room.assert(`paper 2003 has RFID`, ["text", "d07911a3"])
room.assert(`paper 2004 has RFID`, ["text", "91b4d108"])
room.assert(`paper 2005 has RFID`, ["text", "53825027"])
room.assert(`paper 2006 has RFID`, ["text", "10af78a3"])
room.assert(`paper 2007 has RFID`, ["text", "7341a727"])
room.assert(`paper 2008 has RFID`, ["text", "2574c72d"])
room.assert(`paper 2009 has RFID`, ["text", "b680cc21"])
room.assert(`paper 2010 has RFID`, ["text", "737e9c27"])
room.assert(`paper 2011 has RFID`, ["text", "4221cd24"])
room.assert(`paper 5 has RFID`, ["text", "d01ff625"])
room.assert(`paper 1100 has RFID`, ["text", "e21eef27"])
room.assert(`paper 1013 has RFID`, ["text", "7bdbe359"])

run();
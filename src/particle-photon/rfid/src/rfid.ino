// #include "SparkJson.h"
#include "HttpClient.h"

/*
 * MFRC522 - Library to use ARDUINO RFID MODULE KIT 13.56 MHZ WITH TAGS SPI W AND R BY COOQROBOT.
 * The library file MFRC522.h has a wealth of useful info. Please read it.
 * The functions are documented in MFRC522.cpp.
 *
 * Based on code Dr.Leong   ( WWW.B2CQSHOP.COM )
 * Created by Miguel Balboa (circuitito.com), Jan, 2012.
 * Rewritten by Søren Thing Andersen (access.thing.dk), fall of 2013 (Translation to English, refactored, comments, anti collision, cascade levels.)
 * Released into the public domain.
 *
 * Sample program showing how to read data from a PICC using a MFRC522 reader on the Arduino SPI interface.
 *----------------------------------------------------------------------------- empty_skull
 * Aggiunti pin per arduino Mega
 * add pin configuration for arduino mega
 * http://mac86project.altervista.org/
 ----------------------------------------------------------------------------- Nicola Coppola
 * Pin layout should be as follows:
 * Signal     Pin              Pin               Pin			Pin
 *            Arduino Uno      Arduino Mega      SPARK			MFRC522 board
 * ---------------------------------------------------------------------------
 * Reset      9                5                 ANY (D2)		RST
 * SPI SS     10               53                ANY (A2)		SDA
 * SPI MOSI   11               51                A5				MOSI
 * SPI MISO   12               50                A4				MISO
 * SPI SCK    13               52                A3				SCK
 *
 * The reader can be found on eBay for around 5 dollars. Search for "mf-rc522" on ebay.com.
 */

//#include <SPI.h>
#include "MFRC522.h"

#define RST_PIN D2

#define SS_PIN D1
#define SS_PIN_B D0
#define SS_PIN_C D3
#define SS_PIN_D D4
#define SS_PIN_E D5

// Create MFRC522 instances
MFRC522 mfrc522(SS_PIN, RST_PIN);
// MFRC522 mfrc522_b(SS_PIN_B, RST_PIN);
// MFRC522 mfrc522_c(SS_PIN_C, RST_PIN);
// MFRC522 mfrc522_d(SS_PIN_D, RST_PIN);
// MFRC522 mfrc522_e(SS_PIN_E, RST_PIN);
// MFRC522 mfrc522_f(SS_PIN_F, RST_PIN);

HttpClient http;

// Headers currently need to be set at init, useful for API keys etc.
http_header_t headers[] = {
    {"Content-Type", "application/json"},
    {"Accept", "application/json"},
    {"Accept", "*/*"},
    {NULL, NULL} // NOTE: Always terminate headers will NULL
};

http_request_t request;
http_response_t response;

String myID = System.deviceID();

unsigned long lastTime = 0;
unsigned long now = 0;

void publishValueMessage(int sensorId, String sensorValue)
{
  char str[300];
  sprintf(str, "{\"claim\":\"Photon%s read \\\"%s\\\" on sensor %i\", \"retract\":\"$ $ Photon%s read $ on sensor %i\"}", (const char *)myID, sensorValue.c_str(), sensorId, (const char *)myID, sensorId);
  Serial.println(str);
  request.ip = {10, 0, 0, 22};
  request.port = 5000;
  request.path = "/cleanup-claim";
  request.body = str;
  Serial.println(request.body);
  http.post(request, response, headers);
  Serial.print("Application>\tResponse status: ");
  Serial.println(response.status);
}

void publishValueMessages(String sensorValue1, String sensorValue2, String sensorValue3, String sensorValue4, String sensorValue5)
{
  char str[400];
  sprintf(str, "{\"claim\":[\"Photon%s read \\\"%s\\\" on sensor %i\", \"Photon%s read \\\"%s\\\" on sensor %i\", \"Photon%s read \\\"%s\\\" on sensor %i\", \"Photon%s read \\\"%s\\\" on sensor %i\", \"Photon%s read \\\"%s\\\" on sensor %i\"], \"retract\":\"$ $ Photon%s read $ on sensor $\"}",
          (const char *)myID, sensorValue1.c_str(), 1,
          (const char *)myID, sensorValue2.c_str(), 2,
          (const char *)myID, sensorValue3.c_str(), 3,
          (const char *)myID, sensorValue4.c_str(), 4,
          (const char *)myID, sensorValue5.c_str(), 5,
          (const char *)myID);
  Serial.println(str);
  request.ip = {10, 0, 0, 22};
  request.port = 5000;
  request.path = "/cleanup-claim";
  request.body = str;
  Serial.println(request.body);
  http.post(request, response, headers);
  Serial.print("Application>\tResponse status: ");
  Serial.println(response.status);
}

void setup()
{
  Serial.begin(9600); // Initialize serial communications with the PC
  mfrc522.setSPIConfig();
  // mfrc522_b.setSPIConfig();
  // mfrc522_c.setSPIConfig();
  // mfrc522_d.setSPIConfig();
  // mfrc522_e.setSPIConfig();

  mfrc522.PCD_Init();   // Init MFRC522 card
  // mfrc522_b.PCD_Init(); // Init MFRC522 card
  // mfrc522_c.PCD_Init(); // Init MFRC522 card
  // mfrc522_d.PCD_Init(); // Init MFRC522 card
  // mfrc522_e.PCD_Init(); // Init MFRC522 card
  Serial.println("Scan PICC to see UID and type...");
}

String check_reader(MFRC522 reader)
{
  byte bufferATQA[2];
  byte bufferSize = sizeof(bufferATQA);
  byte result = reader.PICC_WakeupA(bufferATQA, &bufferSize);
  if (!(result == MFRC522::STATUS_OK || result == MFRC522::STATUS_COLLISION))
  {
    Serial.println("STATUS is not OK or COLLISION");
    return "null";
  }

  // Select one of the cards
  if (!reader.PICC_ReadCardSerial())
  {
    return "null";
  }

  MFRC522::Uid *uid = &(reader.uid);
  String cardUidString = "";
  Serial.print("Card UID: ");
  for (byte i = 0; i < uid->size; i++)
  {
    cardUidString = String(cardUidString + String(uid->uidByte[i] < 0x10 ? "0" : ""));
    cardUidString = String(cardUidString + String(uid->uidByte[i], HEX));
  }
  Serial.println(cardUidString);

  reader.PICC_HaltA();
  return cardUidString;
}

void loop()
{
  
  delay(50);

  lastTime = millis();  

  String val_a = check_reader(mfrc522);
  // String val_b = check_reader(mfrc522_b);
  // String val_c = check_reader(mfrc522_c);
  // String val_d = check_reader(mfrc522_d);
  // String val_e = check_reader(mfrc522_e);
  // String val_f = check_reader(mfrc522_f);

  now = millis();
  Serial.printlnf("rfid read lag: %lu ms", (now - lastTime));
  lastTime = millis();

  // publishValueMessages(val_a, val_b, val_c, val_d, val_e);
  publishValueMessage(1, val_a);
  // publishValueMessage(2, val_b);
  // publishValueMessage(3, val_c);
  // publishValueMessage(4, val_d);
  // publishValueMessage(5, val_e);
  // publishValueMessage(6, val_f);

  now = millis();
  Serial.printlnf("send lag: %lu ms", (now-lastTime));
}
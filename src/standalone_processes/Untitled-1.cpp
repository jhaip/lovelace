#include "HttpClient.h"

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

void publishValueMessage(char body[])
{
    request.ip = {10, 0, 0, 185};
    request.port = 5000;
    request.path = "/cleanup-claim";
    request.body = str;
    Serial.println(request.body);
    http.post(request, response, headers);
    Serial.print("Application>\tResponse status: ");
    Serial.println(response.status);
}
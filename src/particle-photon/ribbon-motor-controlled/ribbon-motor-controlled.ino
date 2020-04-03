int directionPin = D0;
int pwmPin = D2;
int dutyCycle = 170;
int onTimeMs = 10;
int delayBetweenTicks = 5000;

void setup()
{
    bool successAttachingSetDelayValue = Particle.function("setDelayValue", setDelayValue);
    pinMode(directionPin, OUTPUT);
    pinMode(pwmPin, OUTPUT);
}

void loop()
{
    digitalWrite(directionPin, LOW);
    analogWrite(pwmPin, dutyCycle);
    delay(onTimeMs);
    analogWrite(pwmPin, 0);
    delay(delayBetweenTicks);
}

int setDelayValue(String data)
{
    Serial.print("Received cloud value:");
    Serial.println(data);
    Serial.println("--");
    delayBetweenTicks = data.toInt();
    return 1;
}
#include "Arduino.h"
#include "doodit.h"

int led1 = 7;
int led2 = 4;

/**
 * Arduino initalisation.
 */
void setup() {
  Serial.begin(9600);
  // initialize the digital pin as an output.
  pinMode(led1, OUTPUT);
  pinMode(led2, OUTPUT);
}


// the loop routine runs over and over again forever:
void loop() {
  if(Serial.available() >= 5) {
    Command c = ReadCommand();
    int pin = c.argument;
  
    switch (c.instruction) {
      case 'u':
        if(pin == 1){
          digitalWrite(led1, HIGH);
        } else if(pin == 2){
          digitalWrite(led2, HIGH);
        }
        break;
      case 'd':
        if(pin == 1){
          digitalWrite(led1, LOW);
        } else if(pin == 2){
          digitalWrite(led2, LOW);
        }

        break;
    }
  }
  delay(10);

}

Command ReadCommand() {
  union {
    char b[4];
    int f;
  } diego;

  // Read the command identifier and argument from the serial port.
  char c = Serial.read();
  Serial.readBytes(diego.b, 4);

  return (Command) {c, diego.f};
}


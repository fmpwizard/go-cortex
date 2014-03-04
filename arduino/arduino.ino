#include "Arduino.h"
#include "doodit.h"

int led1 = 7;
int led2 = 4;
int led3 = 13;
int led4 = 11;
int led5 = 8;
int led6 = 2;

/**
 * Arduino initalisation.
 */
void setup() {
  Serial.begin(9600);
  // initialize the digital pin as an output.
  pinMode(led1, OUTPUT);
  pinMode(led2, OUTPUT);
  pinMode(led3, OUTPUT);
  pinMode(led4, OUTPUT);
  pinMode(led5, OUTPUT);
  pinMode(led6, OUTPUT);
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
        } else if(pin == 3){
          digitalWrite(led3, HIGH);
        } else if(pin == 4){
          digitalWrite(led4, HIGH);
        } else if(pin == 5){
          digitalWrite(led5, HIGH);
        } else if(pin == 6){
          digitalWrite(led6, HIGH);
        }
        
        break;
      case 'd':
        if(pin == 1){
          digitalWrite(led1, LOW);
        } else if(pin == 2){
          digitalWrite(led2, LOW);
        } else if(pin == 3){
          digitalWrite(led3, LOW);
        } else if(pin == 4){
          digitalWrite(led4, LOW);
        } else if(pin == 5){
          digitalWrite(led5, LOW);
        } else if(pin == 6){
          digitalWrite(led6, LOW);
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


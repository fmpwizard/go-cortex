/*
 HC-SR04 Ping distance sensor:
 VCC to arduino 5v 
 GND to arduino GND
 Echo to Arduino pin 5 
 Trig to Arduino pin 6
 
 This sketch originates from Virtualmix: http://goo.gl/kJ8Gl
 Has been modified by Winkle ink here: http://winkleink.blogspot.com.au/2012/05/arduino-hc-sr04-ultrasonic-distance.html
 And modified further by ScottC here: http://arduinobasics.blogspot.com.au/2012/11/arduinobasics-hc-sr04-ultrasonic-sensor.html
 on 10 Nov 2012.
 
 This is a mix from several sources, as this arduino has two functions:
 
 1- Has a ultrasonic distance sensor used to signal Go Cortex that it should start recording a voice command
 2- Based on the command processed by Wit, it will turn on/off any of the 6 LEDs connected to it.
 */

#include "Arduino.h"
#include "doodit.h"

int led1 = 7;
int led2 = 4;
int led3 = 13;
int led4 = 11;
int led5 = 8;
int led6 = 2;

int echoPin = 5; // Echo Pin
int trigPin = 6; // Trigger Pin
int LEDPin = 12; // Listening for voice LED

int maximumRange = 200; // Maximum range needed
int minimumRange = 0; // Minimum range needed
long duration, distance; // Duration used to calculate distance
boolean sent = false; // did we just sent a signal to start recording? 


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
  
  pinMode(trigPin, OUTPUT);
  pinMode(echoPin, INPUT);
  pinMode(LEDPin, OUTPUT); // Use LED indicator
}

void ultrasonicSensor() {
  //digitalWrite(LEDPin, LOW);
  /* The following trigPin/echoPin cycle is used to determine the
  distance of the nearest object by bouncing soundwaves off of it. */ 
  digitalWrite(trigPin, LOW); 
  delayMicroseconds(2); 

  digitalWrite(trigPin, HIGH);
  delayMicroseconds(10); 
 
  digitalWrite(trigPin, LOW);
  duration = pulseIn(echoPin, HIGH);
 
  //Calculate the distance (in cm) based on the speed of sound.
  distance = duration/58.2;
 
  if (distance >= maximumRange || distance <= minimumRange){
    digitalWrite(LEDPin, LOW);
    sent = false;
  } else {
    if (sent == false){
      Serial.println("1");
      sent = true;
      digitalWrite(LEDPin, HIGH); 
    }
    
  }
}

// the loop routine runs over and over again forever:
void loop() {
  ultrasonicSensor();
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
  delay(50);

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


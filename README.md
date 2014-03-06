#Go Cortex

`go-cortex` is a service that listens for sentences and tries to understand what you meant and goes off
to do what you asked for.

In the background it uses [Wit.ai](https://wit.ai/) to process the text and get back an `intent` with some parameters extracted from the sentence.

## Example

At home I have a raspberrypi connected to an Arduino board with 6 LEDs. I can go on my browser to `http://raspberrypi:8080/?q=turn+light+6+on` and cortex will send the text `turn light 6 on` to wit and get back the intent `light`, with two parameters, one is the number `6` and the other is `on`.

The beauty here is that I can write different kinds of sentences that mean the same, and wit will do the heavy work of trying to understand them.

I can say `The light 5 should really be on`, and wit will know what I meant, and in return cortext will process the command and turn the LED number 5 on.

## Setup

To allow your user to send data to the usb connected Arduino, you will have to add your current
user to the group `dialout`.

On debian this is: (I'm using a raspberry pi here)

```
sudo usermod -a -G dialout pi
```

Once you do that, you can either logout and login again, or you can start a new session with the new group by doing:

```
su - pi
```

You will also need a wit account, please refer to their [site](https://wit.ai/) for moree information. For now, access to their service is free as long as you are ok sharing your intents/data with them.

Once your wit account is activated, take your time to follow their tour and create some intentions.

### Optional

If you want to control an Arduino, then you will need to set on up, I'll post diagrams for the simple one I have running at home.

There is an arduino folder in this repo that has two files you need to compile and send to the arduino board.

## Running cortex

Assuming you already have `go` installed and have `$GOPATH` setup, then type:

```
go get github.com/fmpwizard/go-cortex
go-cortex --witAccessToken=<token here with no quotes> -httpPort=8080 
```

and you are ready, if you are running this locally, go to `http://127.0.0.1:8080?q=<some command here>` and see the magic


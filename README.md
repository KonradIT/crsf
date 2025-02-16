# CRSF parser

A dead simple CRSF (eg: ELRS packets) parser.

Tested with Happymodel ELRS 915MHz RX using ELRS v3.5.3 + MILBETA + MafiaELRS.

## Features:

- Reads from serial connection
- Interprets control bits, up to 16 channels supported.

## Usage:

```go
// Get new instance:
port := "/dev/ttyUSB0"
baudrate := 425000
timeout := 1 * time.Second
instance := crsf.New(port, baudrate, timeout)

// Open serial connection:
err := instance.Start()
if err != nil {
    fmt.Println(err)
    return
}

// dont forget to close:
defer instance.Close()

// Read data:
instance.Parse(func(packet packet.Packet) {
	fmt.Printf("packet: %v\n", packet.Channels)
})

```

Read from a file:

```bash
$ ./parse-raw -i data-from-fc.bin
Content: {
  "header": {
    "sync_byte": 200,
    "frame_length": 24,
    "type": 22
  },
  "payload": "28PeLb4HcPCBD+DgAx/4KAgAAEx84g==",
  "crc": 255
}
Channels:
[987 984 183 991 1792 992 992 1792 992 992 992 1044 0 0 1811 1811]
```

## Use a ELRS transmitter as a joystick on Windows:

```bash
$ .\examples\joystick\build-exe.bat
$ .\joystick.exe -p COM10
```

**Wiring is as follows:**

![](https://i.imgur.com/3eWjbXS.jpeg)

- ELRS TX -> FTDI RX
- ELRS RX -> FTDI TX
- ELRS 5V -> FTDI VCC
- ELRS GND -> FTDI GND

I notice a better latency when using this setup than when using the ELRS BLE functionality.
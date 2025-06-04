package sensor

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"time"
)

type Bus interface {
	Tx(w, r []byte) error
}

type Sensor interface {
	Init() error
	Stop() error
	Clean() error
	Read() (Measurement, error)
	IsMeasuring() (bool, error)
}

type SCD4XSensor struct {
	bus Bus
}

type Measurement struct {
	CO2         float32
	Temperature float32
	Humidity    float32
}

const (
	cmdGetDataReadyStatus       = 0xe4b8
	cmdReadMeasurement          = 0xec05
	cmdSetAmbientPressure       = 0xe000
	cmdSetTemperatureOffset     = 0x241d
	cmdStartPeriodicMeasurement = 0x21b1
	cmdStopPeriodicMeasurement  = 0x3f86
)

func New(bus Bus) *SCD4XSensor {
	return &SCD4XSensor{bus: bus}
}

func (s *SCD4XSensor) Init() error {
	slog.Info("Sending start measurement command to SCD4X")
	return s.sendCommand(cmdStartPeriodicMeasurement, nil)
}

func (s *SCD4XSensor) Stop() error {
	if err := s.sendCommand(cmdStopPeriodicMeasurement, nil); err != nil {
		return err
	}
	// According to the datasheet, the sensor will respond to other commands
	// only 500 ms after stop_periodic_measurement has been issued.
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (s *SCD4XSensor) Clean() error {
	return nil
}

func (s *SCD4XSensor) Read() (Measurement, error) {
	if err := s.sendCommand(cmdReadMeasurement, nil); err != nil {
		return Measurement{}, fmt.Errorf("sendCommand error: %w", err)
	}
	time.Sleep(2 * time.Millisecond) // Обычно достаточно

	buf := make([]byte, 9)
	if err := s.bus.Tx(nil, buf); err != nil {
		return Measurement{}, fmt.Errorf("I2C Tx error: %w", err)
	}

	for i := 0; i < 9; i += 3 {
		if !validCRC(buf[i:i+2], buf[i+2]) {
			return Measurement{}, fmt.Errorf("CRC check failed at position %d", i)
		}
	}

	co2Raw := binary.BigEndian.Uint16(buf[0:2])
	tempRaw := binary.BigEndian.Uint16(buf[3:5])
	humRaw := binary.BigEndian.Uint16(buf[6:8])

	co2 := float32(co2Raw)
	temperature := float32(-45.0 + 175.0*float32(tempRaw)/65535.0)
	humidity := float32(100.0 * float32(humRaw) / 65535.0)

	return Measurement{
		CO2:         co2,
		Temperature: temperature,
		Humidity:    humidity,
	}, nil
}

func (s *SCD4XSensor) sendCommand(cmd uint16, args []byte) error {
	if args != nil && len(args)%2 != 0 {
		return fmt.Errorf("arguments length must be even (pairs of bytes)")
	}

	buf := []byte{byte(cmd >> 8), byte(cmd & 0xFF)}
	if args != nil {
		for i := 0; i < len(args); i += 2 {
			chunk := args[i : i+2]
			crc := calcCRC(chunk)
			buf = append(buf, chunk[0], chunk[1], crc)
		}
	}

	for attempt := 1; attempt <= 3; attempt++ {
		if err := s.bus.Tx(buf, nil); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("I2C Tx failed after retries")
}

func (s *SCD4XSensor) IsMeasuring() (bool, error) {
	if err := s.sendCommand(cmdGetDataReadyStatus, nil); err != nil {
		return false, err
	}
	time.Sleep(3 * time.Millisecond)

	buf := make([]byte, 3)
	if err := s.bus.Tx(nil, buf); err != nil {
		return false, err
	}
	if !validCRC(buf[:2], buf[2]) {
		return false, fmt.Errorf("CRC error")
	}
	status := binary.BigEndian.Uint16(buf[:2])
	dataReady := (status & 0x07FF) != 0
	return dataReady, nil
}

func validCRC(data []byte, crc byte) bool {
	return calcCRC(data) == crc
}

func calcCRC(data []byte) byte {
	crc := byte(0xFF)
	for _, b := range data {
		crc ^= b
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x31
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

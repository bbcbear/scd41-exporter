package main

import (
	"fmt"
	"log"
	
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/scd4x"
	"periph.io/x/host/v3"
)

func main() {
	//// init start
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Открытие I2C
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	dev, err := scd4x.NewI2C(bus, scd4x.SensorAddress)
	if err != nil {
		log.Fatal(err)
	}
	//// init end

	//// set pressure start
	cfg, err := dev.GetConfiguration()
	if err == nil {
		fmt.Printf("Configuration: %#v\n", cfg)
	} else {
		fmt.Println(err)
	}

	var p physic.Pressure
	if err := p.Set("88557Pa"); err != nil {
		log.Fatal(err)
	}
	cfg.AmbientPressure = p

	if err := dev.SetConfiguration(cfg); err != nil {
		log.Fatal(err)
	}
	fmt.Println("pressure is set")
	//// set pressure end

	/// get data start
	env := scd4x.Env{}
	for {
		err = dev.Sense(&env)
		if err == nil {
			fmt.Println(env.String())
		} else {
			fmt.Println(err)
		}
	}
	/// get data end

	///  SenseContinuous 
	///  func (d *Dev) SenseContinuous(interval time.Duration) (<-chan Env, error) {
}

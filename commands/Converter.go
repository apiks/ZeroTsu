package commands

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/r-anime/ZeroTsu/config"
)

// Converts measurement units to the requested one
func ConverterHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if strings.HasPrefix(m.Content, config.BotPrefix) {

		if m.Author.ID == config.BotID {
			return
		}

		// Puts the command to lowercase and then splits every word (so you can check if it's the proper number of strings
		messageLowercase := strings.ToLower(m.Content)

		//Checks if BotPrefix + convert was used
		if strings.HasPrefix(messageLowercase, config.BotPrefix+"convert ") && (messageLowercase != (config.BotPrefix + "convert")) {

			// Splits the original message by spaces
			messageSplit := strings.Split(messageLowercase, " ")

			if len(messageSplit) > 3 {
				// Converts from Celsius to Fahrenheit or Kelvin
				if ((messageSplit[2] == "c" || messageSplit[2] == "celsius") && messageSplit[2] != "cm") ||
					((strings.Contains(messageSplit[1], "c") && strings.Contains(messageSplit[1], "cm") == false && strings.Contains(messageSplit[1], "cms") == false &&

						strings.Contains(messageSplit[1], "centimeter") == false && strings.Contains(messageSplit[1], "centimeters") == false &&
						strings.Contains(messageSplit[1], "centimetre") == false && strings.Contains(messageSplit[1], "centimetres") == false &&
						strings.Contains(messageSplit[1], "inch") == false && strings.Contains(messageSplit[1], "inches") == false) ||

						strings.Contains(messageSplit[1], "celsius")) {

					CelsiusHandler(messageSplit, *s, *m)

					// Converts from Fahrenheit to Celsius or Kelvin
				} else if (messageSplit[2] == "f" || messageSplit[2] == "fahrenheit") ||
					((strings.Contains(messageSplit[1], "f") || strings.Contains(messageSplit[1], "fahrenheit")) &&

						(strings.Contains(messageSplit[1], "ft") == false && strings.Contains(messageSplit[1], "feet") == false &&
							strings.Contains(messageSplit[1], "foot") == false)) {

					FahrenheitHandler(messageSplit, *s, *m)

					// Converts from Kelvin to Celsius or Fahrenheit
				} else if (messageSplit[2] == "k" || messageSplit[2] == "kelvin" || messageSplit[2] == "kalvin") ||

					((strings.Contains(messageSplit[1], "k") && strings.Contains(messageSplit[1], "km") == false && strings.Contains(messageSplit[1], "kms") == false &&
						strings.Contains(messageSplit[1], "kilometer") == false && strings.Contains(messageSplit[1], "kilometers") == false &&
						strings.Contains(messageSplit[1], "kilometre") == false && strings.Contains(messageSplit[1], "kilometres") == false) ||

						strings.Contains(messageSplit[1], "kelvin") || strings.Contains(messageSplit[1], "kalvin")) {

					KelvinHandler(messageSplit, *s, *m)

					// Converts from Centimeter to Millimeter, Inch, Meter, Foot, Kilometer or Mile
				} else if (messageSplit[2] == "cm" || messageSplit[2] == "cms" || messageSplit[2] == "centimeter" || messageSplit[2] == "centimeters" ||
					messageSplit[2] == "centimetre" || messageSplit[2] == "centimetres") ||

					(strings.Contains(messageSplit[1], "cm") || strings.Contains(messageSplit[1], "cms") || strings.Contains(messageSplit[1], "centimeter") ||
						strings.Contains(messageSplit[1], "centimeters") || strings.Contains(messageSplit[1], "centimetre") || strings.Contains(messageSplit[1], "centimetres")) {

					CentimeterHandler(messageSplit, *s, *m)

					// Converts from Millimeter to Centimeter, Inch, Meter, Foot, Kilometer or Mile
				} else if (messageSplit[2] == "mm" || messageSplit[2] == "mms" || messageSplit[2] == "millimeter" || messageSplit[2] == "millimeters" ||
					messageSplit[2] == "millimetre" || messageSplit[2] == "millimetres") ||

					(strings.Contains(messageSplit[1], "mm") || strings.Contains(messageSplit[1], "mms") || strings.Contains(messageSplit[1], "millimeter") ||
						strings.Contains(messageSplit[1], "millimeters") || strings.Contains(messageSplit[1], "millimetre") || strings.Contains(messageSplit[1], "millimetres")) {

					MillimeterHandler(messageSplit, *s, *m)

					// Converts from Inch to Millimeter, Centimeter, Meter, Foot, Kilometer or Mile
				} else if (messageSplit[2] == "inch" || messageSplit[2] == "inches" || messageSplit[2] == "'" ||
					(messageSplit[2] == "in" && strings.Contains(messageSplit[1], "'") == false) || messageSplit[2] == "ins") ||

					(strings.Contains(messageSplit[1], "inch") || strings.Contains(messageSplit[1], "inches") ||
						((strings.Contains(messageSplit[1], "'") && strings.Contains(messageSplit[1], "\"") == false) &&
							strings.Count(messageSplit[1], "'") == 1) ||
						strings.Contains(messageSplit[1], "in") || strings.Contains(messageSplit[1], "ins")) {

					InchHandler(messageSplit, *s, *m)

					// Converts from Foot to Millimeter, Centimeter, Meter, Inch, Kilometer or Mile
				} else if (messageSplit[2] == "foot" || messageSplit[2] == "feet" || messageSplit[2] == "''" || messageSplit[2] == "ft" ||
					messageSplit[2] == "\"") ||

					(strings.Contains(messageSplit[1], "foot") || strings.Contains(messageSplit[1], "feet") ||
						(strings.Contains(messageSplit[1], "''") && strings.Count(messageSplit[1], "'") < 3) ||
						(strings.Contains(messageSplit[1], "\"") && strings.Contains(messageSplit[1], "'") == false) ||
						strings.Contains(messageSplit[1], "ft")) {

					FootHandler(messageSplit, *s, *m)

					// Converts from Mile to Millimeter, Centimeter, Foot, Inch, Kilometer or Meter
				} else if (messageSplit[2] == "mile" || messageSplit[2] == "miles" || messageSplit[2] == "mi") ||
					(strings.Contains(messageSplit[1], "mile") || strings.Contains(messageSplit[1], "miles") || strings.Contains(messageSplit[1], "mi")) {

					MileHandler(messageSplit, *s, *m)

					// Converts from Kilometer to Millimeter, Centimeter, Foot, Inch, Meter or Mile
				} else if (messageSplit[2] == "km" || messageSplit[2] == "kms" || messageSplit[2] == "kilometer" || messageSplit[2] == "kilometers" ||
					messageSplit[2] == "kilometre" || messageSplit[2] == "kilometres") ||

					(strings.Contains(messageSplit[1], "km") || strings.Contains(messageSplit[1], "kms") || strings.Contains(messageSplit[1], "kilometer") ||
						strings.Contains(messageSplit[1], "kilometers") || strings.Contains(messageSplit[1], "kilometre") || strings.Contains(messageSplit[1], "kilometres")) {

					KilometerHandler(messageSplit, *s, *m)

					// Converts from Meter to Millimeter, Centimeter, Foot, Inch, Kilometer or Mile
				} else if (messageSplit[2] == "meter" || messageSplit[2] == "meters" || messageSplit[2] == "m" || messageSplit[2] == "metre" ||
					messageSplit[2] == "metres") ||

					// CAREFUL WITH "m" CONTAINS FOR FUTURE ADDITIONS
					(strings.Contains(messageSplit[1], "meter") || strings.Contains(messageSplit[1], "meters") ||
						strings.Contains(messageSplit[1], "m") || strings.Contains(messageSplit[1], "metre") || strings.Contains(messageSplit[1], "metres")) {

					MeterHandler(messageSplit, *s, *m)

					//Parses and converts Inch Height (e.g. 5"5')
				} else {

					var heightBool bool

					// Regex checks if a height was used
					re := regexp.MustCompile("[0-9]+[\"']+[0-9]+'")
					heightCheck := re.FindAllString(messageLowercase, 1)

					if heightCheck != nil {

						heightBool = true
					}

					if heightBool == true {

						var cmString string

						// Splits the height from """ and "'"
						r := regexp.MustCompile("[\"']")

						split := r.Split(messageSplit[1], 3)

						// Calculates the total cm number before printing
						foot, err := strconv.ParseFloat(split[0], 64)
						if err != nil {

							fmt.Println("Error: ", err)
						}
						inch, err := strconv.ParseFloat(split[1], 64)
						if err != nil {

							fmt.Println("Error: ", err)
						}
						cm := (foot * 30.48) + (inch * 2.54)

						// If the result is a whole number it returns that, else the float
						if isFloatInt(cm) == true {

							fahrenheitNumInt := int(cm)

							cmString = strconv.Itoa(fahrenheitNumInt)
						} else {

							cmString = strconv.FormatFloat(cm, 'f', 2, 64)
						}

						// Prints result
						_, err = s.ChannelMessageSend(m.ChannelID, messageSplit[1]+" is "+cmString+"cm")
						if err != nil {

							fmt.Println("Error: ", err)
						}

					}
				}
			}
		}
	}
}

// Checks if the float number is a whole number or not and returns bool
func isFloatInt(floatValue float64) bool {
	return math.Mod(floatValue, 1.0) == 0
}

// Takes in a celsius number and converts to fahrenheit
func CelsiusToFahrenheit(celsiusNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		fahrenheitString string
		celsiusString    string
	)

	// Celsius to Fahrenheit formula
	fahrenheitNum := (celsiusNum * 1.8) + 32

	// If the result is a whole number it returns that, else the float
	if isFloatInt(fahrenheitNum) == true {

		fahrenheitNumInt := int(fahrenheitNum)

		fahrenheitString = strconv.Itoa(fahrenheitNumInt)
	} else {

		fahrenheitString = strconv.FormatFloat(fahrenheitNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(celsiusNum) == true {

		celsiusNumInt := int(celsiusNum)

		celsiusString = strconv.Itoa(celsiusNumInt)
	} else {

		celsiusString = strconv.FormatFloat(celsiusNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, celsiusString+"C is "+fahrenheitString+"F")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a celsius number and converts to kelvin
func CelsiusToKelvin(celsiusNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		kelvinString  string
		celsiusString string
	)

	// Celsius to Fahrenheit formula
	kelvinNum := celsiusNum + 273.15

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kelvinNum) == true {

		kelvinNumInt := int(kelvinNum)

		kelvinString = strconv.Itoa(kelvinNumInt)
	} else {

		kelvinString = strconv.FormatFloat(kelvinNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(celsiusNum) == true {

		celsiusNumInt := int(celsiusNum)

		celsiusString = strconv.Itoa(celsiusNumInt)
	} else {

		celsiusString = strconv.FormatFloat(celsiusNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, celsiusString+"C is "+kelvinString+"K")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Celsius method
func CelsiusHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper celsius number
	celsiusString := messageSplit[1]

	// Parses the celsius number correctly
	if strings.Contains(celsiusString, "celsius") {

		celsiusString = strings.Replace(celsiusString, "celsius", "", 1)

	} else if strings.Contains(celsiusString, "c") && strings.Contains(celsiusString, "cm") == false {

		celsiusString = strings.Replace(celsiusString, "c", "", 1)
	}

	// Converts celsius number to a float64 and prints error if it cannot
	celsiusNum, err := strconv.ParseFloat(celsiusString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire celsius blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "f" || messageSplit[3] == "fahrenheit" {

				// Initiates the Celsius to Fahrenheit Conversion
				CelsiusToFahrenheit(celsiusNum, s, m)
			} else if messageSplit[3] == "k" || messageSplit[3] == "kelvin" || messageSplit[3] == "kalvin" {

				// Initiates the Celsius to Kelvin Conversion
				CelsiusToKelvin(celsiusNum, s, m)
			}

		} else if messageSplit[4] == "f" || messageSplit[4] == "fahrenheit" {

			// Initiates the Celsius to Fahrenheit Conversion
			CelsiusToFahrenheit(celsiusNum, s, m)
		} else if messageSplit[4] == "k" || messageSplit[4] == "kelvin" || messageSplit[4] == "kalvin" {

			// Initiates the Celsius to Kelvin Conversion
			CelsiusToKelvin(celsiusNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

///////////////////

// Takes in a fahrenheit number and converts to celsius
func FahrenheitToCelsius(fahrenheitNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		fahrenheitString string
		celsiusString    string
	)

	// Celsius to Fahrenheit formula
	CelsiusNum := (fahrenheitNum - 32) * 0.5556

	// If the result is a whole number it returns that, else the float
	if isFloatInt(CelsiusNum) == true {

		celsiusNumInt := int(CelsiusNum)

		celsiusString = strconv.Itoa(celsiusNumInt)
	} else {

		celsiusString = strconv.FormatFloat(CelsiusNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(fahrenheitNum) == true {

		fahrenheitNumInt := int(fahrenheitNum)

		fahrenheitString = strconv.Itoa(fahrenheitNumInt)
	} else {

		fahrenheitString = strconv.FormatFloat(fahrenheitNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, fahrenheitString+"F is "+celsiusString+"C")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a fahrenheit number and converts to kelvin
func FahrenheitToKelvin(fahrenheitNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		fahrenheitString string
		kelvinString     string
	)

	// Fahrenheit to Kelvin formula
	kelvinNum := (fahrenheitNum + 459.67) * 0.5556

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kelvinNum) == true {

		kelvinNumInt := int(kelvinNum)

		kelvinString = strconv.Itoa(kelvinNumInt)
	} else {

		kelvinString = strconv.FormatFloat(kelvinNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(fahrenheitNum) == true {

		fahrenheitNumInt := int(fahrenheitNum)

		fahrenheitString = strconv.Itoa(fahrenheitNumInt)
	} else {

		fahrenheitString = strconv.FormatFloat(fahrenheitNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, fahrenheitString+"F is "+kelvinString+"K")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Fahrenheit method
func FahrenheitHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper fahrenheit number
	fahrenheitString := messageSplit[1]

	// Parses the fahrenheit number correctly
	if strings.Contains(fahrenheitString, "fahrenheit") {

		fahrenheitString = strings.Replace(fahrenheitString, "fahrenheit", "", 1)

	} else if strings.Contains(fahrenheitString, "f") {

		fahrenheitString = strings.Replace(fahrenheitString, "f", "", 1)
	}

	// Converts fahrenheit number to a float64 and prints error if it cannot
	fahrenheitNum, err := strconv.ParseFloat(fahrenheitString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire fahrenheit blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "c" || messageSplit[3] == "celsius" {

				// Initiates the Fahrenheit to Celsius Conversion
				FahrenheitToCelsius(fahrenheitNum, s, m)
			} else if messageSplit[3] == "k" || messageSplit[3] == "kelvin" || messageSplit[3] == "kalvin" {

				// Initiates the Fahrenheit to Kelvin Conversion
				FahrenheitToKelvin(fahrenheitNum, s, m)
			}

		} else if messageSplit[4] == "f" || messageSplit[4] == "fahrenheit" {

			// Initiates the Fahrenheit to Celsius Conversion
			FahrenheitToCelsius(fahrenheitNum, s, m)
		} else if messageSplit[4] == "k" || messageSplit[4] == "kelvin" || messageSplit[4] == "kalvin" {

			// Initiates the Fahrenheit to Kelvin Conversion
			FahrenheitToKelvin(fahrenheitNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

///////////////////

// Takes in a kelvin number and converts to celsius
func KelvinToCelsius(kelvinNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		celsiusString string
		kelvinString  string
	)

	// Kelvin to Fahrenheit formula
	celsiusNum := (kelvinNum - 273) * 0.5556

	// If the result is a whole number it returns that, else the float
	if isFloatInt(celsiusNum) == true {

		celsiusNumInt := int(celsiusNum)

		celsiusString = strconv.Itoa(celsiusNumInt)
	} else {

		celsiusString = strconv.FormatFloat(celsiusNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kelvinNum) == true {

		kelvinNumInt := int(kelvinNum)

		kelvinString = strconv.Itoa(kelvinNumInt)
	} else {

		kelvinString = strconv.FormatFloat(kelvinNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, kelvinString+"K is "+celsiusString+"C")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a kelvin number and converts to fahrenheit
func KelvinToFahrenheit(kelvinNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		fahrenheitString string
		kelvinString     string
	)

	// Kelvin to Fahrenheit formula
	fahrenheitNum := (kelvinNum - 273) * 0.5556

	// If the result is a whole number it returns that, else the float
	if isFloatInt(fahrenheitNum) == true {

		fahrenheitNumInt := int(fahrenheitNum)

		fahrenheitString = strconv.Itoa(fahrenheitNumInt)
	} else {

		fahrenheitString = strconv.FormatFloat(fahrenheitNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kelvinNum) == true {

		kelvinNumInt := int(kelvinNum)

		kelvinString = strconv.Itoa(kelvinNumInt)
	} else {

		kelvinString = strconv.FormatFloat(kelvinNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, kelvinString+"K is "+fahrenheitString+"F")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Kelvin method
func KelvinHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper kelvin number
	kelvinString := messageSplit[1]

	// Parses the kelvin number correctly
	if strings.Contains(kelvinString, "kelvin") {

		kelvinString = strings.Replace(kelvinString, "kelvin", "", 1)

	} else if strings.Contains(kelvinString, "k") && strings.Contains(kelvinString, "km") == false && strings.Contains(kelvinString, "kms") == false &&
		strings.Contains(kelvinString, "kilometer") == false && strings.Contains(kelvinString, "kilometers") == false &&
		strings.Contains(kelvinString, "kilometre") == false && strings.Contains(kelvinString, "kilometres") == false {

		kelvinString = strings.Replace(kelvinString, "k", "", 1)

	} else if strings.Contains(kelvinString, "kalvin") {

		kelvinString = strings.Replace(kelvinString, "kalvin", "", 1)
	}

	// Converts kelvin number to a float64 and prints error if it cannot
	kelvinNum, err := strconv.ParseFloat(kelvinString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire kelvin blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "c" || messageSplit[3] == "celsius" {

				// Initiates the Kelvin to Celsius Conversion
				KelvinToCelsius(kelvinNum, s, m)
			} else if messageSplit[3] == "f" || messageSplit[3] == "fahrenheit" {

				// Initiates the Kelvin to Fahrenheit Conversion
				KelvinToFahrenheit(kelvinNum, s, m)
			}

		} else if messageSplit[4] == "c" || messageSplit[4] == "celsius" {

			// Initiates the Kelvin to Celsius Conversion
			KelvinToCelsius(kelvinNum, s, m)
		} else if messageSplit[4] == "f" || messageSplit[4] == "fahrenheit" {

			// Initiates the Kelvin to Fahrenheit Conversion
			KelvinToFahrenheit(kelvinNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

///////////////////

// Takes in a Centimeter number and converts to Millimeter
func CentimeterToMillimeter(centimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		millimeterString string
	)

	// Centimeter to Millimeter formula
	millimeterNum := centimeterNum * 10

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"cm is "+millimeterString+"mm")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a Centimeter number and converts to Inch
func CentimeterToInch(centimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		inchString       string
	)

	// Centimeter to Inch formula
	inchNum := centimeterNum / 2.54

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', 2, 64)
	}

	// Checks if it's 1 inch for grammatically correct output purposes
	if inchNum > 0 && inchNum < 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"cm is "+inchString+" inch")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"cm is "+inchString+" inches")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Centimeter number and converts to Meter
func CentimeterToMeter(centimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		meterString      string
	)

	// Centimeter to Meter formula
	meterNum := centimeterNum / 100

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', 2, 64)
	}

	// Checks if it's 1 meter for grammatically correct output purposes
	if meterNum > 0 && meterNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"cm is "+meterString+" meter")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"cm is "+meterString+" meters")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Centimeter number and converts to Foot
func CentimeterToFoot(centimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		feetString       string
	)

	// Centimeter to Foot formula
	feetNum := centimeterNum / 30.48

	// If the result is a whole number it returns that, else the float
	if isFloatInt(feetNum) == true {

		feetNumInt := int(feetNum)

		feetString = strconv.Itoa(feetNumInt)
	} else {

		feetString = strconv.FormatFloat(feetNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', 2, 64)
	}

	// Checks if it's 1 foot for grammatically correct output purposes
	if feetNum > 0 && feetNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"cm is "+feetString+" foot")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"cm is "+feetString+" feet")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Centimeter number and converts to Kilometer
func CentimeterToKilometer(centimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		kilometerString  string
	)

	// Centimeter to Kilometer formula
	kilometerNum := centimeterNum / 100000

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', -1, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"km is "+kilometerString+"cm")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a Centimeter number and converts to Mile
func CentimeterToMile(centimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		mileString       string
	)

	// Centimeter to mile formula
	mileNum := centimeterNum / 160934.4

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', -1, 64)
	}

	// Checks if it's 1 mile for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"cm is "+mileString+" mile")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"cm is "+mileString+" miles")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Centimeter method
func CentimeterHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper centimeter number
	centimeterString := messageSplit[1]

	// Parses the centimeter number correctly
	if strings.Contains(centimeterString, "centimeters") {

		centimeterString = strings.Replace(centimeterString, "centimeters", "", 1)

	} else if strings.Contains(centimeterString, "centimetres") {

		centimeterString = strings.Replace(centimeterString, "centimetres", "", 1)

	} else if strings.Contains(centimeterString, "centimetre") {

		centimeterString = strings.Replace(centimeterString, "centimetre", "", 1)
	} else if strings.Contains(centimeterString, "centimeter") {

		centimeterString = strings.Replace(centimeterString, "centimeter", "", 1)
	} else if strings.Contains(centimeterString, "cms") {

		centimeterString = strings.Replace(centimeterString, "cms", "", 1)
	} else if strings.Contains(centimeterString, "cm") {

		centimeterString = strings.Replace(centimeterString, "cm", "", 1)
	}

	// Converts centimeter number to a float64 and prints error if it cannot
	centimeterNum, err := strconv.ParseFloat(centimeterString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire centimeter blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "mm" || messageSplit[3] == "mms" || messageSplit[3] == "millimetre" || messageSplit[3] == "millimetres" ||
				messageSplit[3] == "millimeter" || messageSplit[3] == "millimeters" {

				// Initiates the Centimeter to Millimeter Conversion
				CentimeterToMillimeter(centimeterNum, s, m)
			} else if messageSplit[3] == "inch" || messageSplit[3] == "inches" ||
				messageSplit[3] == "'" || messageSplit[3] == "in" || messageSplit[3] == "ins" {

				// Initiates the Centimeter to Inch Conversion
				CentimeterToInch(centimeterNum, s, m)
			} else if messageSplit[3] == "meter" || messageSplit[3] == "meters" ||
				messageSplit[3] == "metre" || messageSplit[3] == "metres" || messageSplit[3] == "m" {

				// Initiates the Centimeter to Meter Conversion
				CentimeterToMeter(centimeterNum, s, m)
			} else if messageSplit[3] == "foot" || messageSplit[3] == "feet" || messageSplit[3] == "\"" || messageSplit[3] == "''" || messageSplit[3] == "ft" {

				// Initiates the Centimeter to Foot Conversion
				CentimeterToFoot(centimeterNum, s, m)
			} else if messageSplit[3] == "kilometer" || messageSplit[3] == "kilometers" ||
				messageSplit[3] == "kilometre" || messageSplit[3] == "kilometres" ||
				messageSplit[3] == "km" || messageSplit[3] == "kms" {

				// Initiates the Centimeter to Kilometer Conversion
				CentimeterToKilometer(centimeterNum, s, m)
			} else if messageSplit[3] == "mile" || messageSplit[3] == "miles" || messageSplit[3] == "mi" {

				// Initiates the Centimeter to Mile Conversion
				CentimeterToMile(centimeterNum, s, m)
			}

		} else if messageSplit[4] == "mm" || messageSplit[4] == "mms" || messageSplit[4] == "millimetre" || messageSplit[4] == "millimetres" ||
			messageSplit[4] == "millimeter" || messageSplit[4] == "millimeters" {

			// Initiates the Centimeter to Millimeter Conversion
			CentimeterToMillimeter(centimeterNum, s, m)
		} else if messageSplit[4] == "inch" || messageSplit[4] == "inches" ||
			messageSplit[4] == "'" || messageSplit[4] == "in" || messageSplit[4] == "ins" {

			// Initiates the Centimeter to Inch Conversion
			CentimeterToInch(centimeterNum, s, m)
		} else if messageSplit[4] == "meter" || messageSplit[4] == "meters" ||
			messageSplit[4] == "metre" || messageSplit[4] == "metres" || messageSplit[3] == "m" {

			// Initiates the Centimeter to Meter Conversion
			CentimeterToMeter(centimeterNum, s, m)
		} else if messageSplit[4] == "foot" || messageSplit[4] == "feet" || messageSplit[4] == "\"" || messageSplit[4] == "''" || messageSplit[3] == "ft" {

			// Initiates the Centimeter to Foot Conversion
			CentimeterToFoot(centimeterNum, s, m)
		} else if messageSplit[4] == "kilometer" || messageSplit[4] == "kilometers" ||
			messageSplit[4] == "kilometre" || messageSplit[4] == "kilometres" ||
			messageSplit[4] == "km" || messageSplit[4] == "kms" {

			// Initiates the Centimeter to Kilometer Conversion
			CentimeterToKilometer(centimeterNum, s, m)
		} else if messageSplit[4] == "mile" || messageSplit[4] == "miles" || messageSplit[4] == "mi" {

			// Initiates the Centimeter to Mile Conversion
			CentimeterToMile(centimeterNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

///////////////////

// Takes in a Millimeter number and converts to Centimeter
func MillimeterToCentimeter(millimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		millimeterString string
	)

	// Millimeter to Centimeter formula
	centimeterNum := millimeterNum / 10

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, millimeterString+"mm is "+centimeterString+"cm")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a Millimeter number and converts to Inch
func MillimeterToInch(millimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		millimeterString string
		inchString       string
	)

	// Millimeter to Inch formula
	inchNum := millimeterNum / 25.4

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// Checks if it's 1 inch for grammatically correct output purposes
	if inchNum > 0 && inchNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, millimeterString+"mm is "+inchString+" inch")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, millimeterString+"mm is "+inchString+" inches")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Millimeter number and converts to Meter
func MillimeterToMeter(millimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		millimeterString string
		meterString      string
	)

	// Millimeter to Meter formula
	meterNum := millimeterNum / 1000

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, millimeterString+"mm is "+meterString+"m")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a Millimeter number and converts to Foot
func MillimeterToFoot(millimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		millimeterString string
		feetString       string
	)

	// Millimeter to Foot formula
	feetNum := millimeterNum / 304.8

	// If the result is a whole number it returns that, else the float
	if isFloatInt(feetNum) == true {

		feetNumInt := int(feetNum)

		feetString = strconv.Itoa(feetNumInt)
	} else {

		feetString = strconv.FormatFloat(feetNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// Checks if it's 1 foot for grammatically correct output purposes
	if feetNum > 0 && feetNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, millimeterString+"mm is "+feetString+" foot")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, millimeterString+"mm is "+feetString+" feet")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Millimeter number and converts to Kilometer
func MillimeterToKilometer(millimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		millimeterString string
		kilometerString  string
	)

	// Millimeter to Kilometer formula
	kilometerNum := millimeterNum / 1000000

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', -1, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, millimeterString+"mm is "+kilometerString+"km")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a Millimeter number and converts to Mile
func MillimeterToMile(millimeterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		millimeterString string
		mileString       string
	)

	// Millimeter to mile formula
	mileNum := millimeterNum / 160934.4

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', -1, 64)
	}

	// Checks if it's 1 mile for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, millimeterString+"mm is "+mileString+" mile")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, millimeterString+"mm is "+mileString+" miles")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Millimeter method
func MillimeterHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper millimeter number
	millimeterString := messageSplit[1]

	// Parses the millimeter number correctly
	if strings.Contains(millimeterString, "millimeters") {

		millimeterString = strings.Replace(millimeterString, "millimeters", "", 1)

	} else if strings.Contains(millimeterString, "millimetres") {

		millimeterString = strings.Replace(millimeterString, "millimetres", "", 1)
	} else if strings.Contains(millimeterString, "millimetre") {

		millimeterString = strings.Replace(millimeterString, "millimetre", "", 1)
	} else if strings.Contains(millimeterString, "millimeter") {

		millimeterString = strings.Replace(millimeterString, "millimeter", "", 1)
	} else if strings.Contains(millimeterString, "mms") {

		millimeterString = strings.Replace(millimeterString, "mms", "", 1)
	} else if strings.Contains(millimeterString, "mm") {

		millimeterString = strings.Replace(millimeterString, "mm", "", 1)
	}

	// Converts millimeter number to a float64 and prints error if it cannot
	millimeterNum, err := strconv.ParseFloat(millimeterString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire millimeter blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "cm" || messageSplit[3] == "cms" || messageSplit[3] == "centimetre" || messageSplit[3] == "centimetres" ||
				messageSplit[3] == "centimeter" || messageSplit[3] == "centimeters" {

				// Initiates the Millimeter to Centimeter Conversion
				MillimeterToCentimeter(millimeterNum, s, m)
			} else if messageSplit[3] == "inch" || messageSplit[3] == "inches" ||
				messageSplit[3] == "'" || messageSplit[3] == "in" || messageSplit[3] == "ins" {

				// Initiates the Millimeter to Inch Conversion
				MillimeterToInch(millimeterNum, s, m)
			} else if messageSplit[3] == "meter" || messageSplit[3] == "meters" ||
				messageSplit[3] == "metre" || messageSplit[3] == "metres" || messageSplit[3] == "m" {

				// Initiates the Millimeter to Meter Conversion
				MillimeterToMeter(millimeterNum, s, m)
			} else if messageSplit[3] == "foot" || messageSplit[3] == "feet" || messageSplit[3] == "\"" || messageSplit[3] == "''" || messageSplit[3] == "ft" {

				// Initiates the Millimeter to Foot Conversion
				MillimeterToFoot(millimeterNum, s, m)
			} else if messageSplit[3] == "kilometer" || messageSplit[3] == "kilometers" ||
				messageSplit[3] == "kilometre" || messageSplit[3] == "kilometres" ||
				messageSplit[3] == "km" || messageSplit[3] == "kms" {

				// Initiates the Millimeter to Kilometer Conversion
				MillimeterToKilometer(millimeterNum, s, m)
			} else if messageSplit[3] == "mile" || messageSplit[3] == "miles" || messageSplit[3] == "mi" {

				// Initiates the Millimeter to Mile Conversion
				MillimeterToMile(millimeterNum, s, m)
			}

		} else if messageSplit[4] == "cm" || messageSplit[4] == "cms" || messageSplit[4] == "centimetre" || messageSplit[4] == "centimetres" ||
			messageSplit[4] == "centimeter" || messageSplit[4] == "centimeters" {

			// Initiates the Millimeter to Centimeter Conversion
			MillimeterToCentimeter(millimeterNum, s, m)
		} else if messageSplit[4] == "inch" || messageSplit[4] == "inches" {

			// Initiates the Millimeter to Centimeter Conversion
			MillimeterToCentimeter(millimeterNum, s, m)
		} else if messageSplit[4] == "inch" || messageSplit[4] == "inches" || messageSplit[4] == "in" || messageSplit[4] == "ins" ||
			messageSplit[4] == "'" {

			// Initiates the Millimeter to Inch Conversion
			MillimeterToInch(millimeterNum, s, m)
		} else if messageSplit[4] == "meter" || messageSplit[4] == "meters" ||
			messageSplit[4] == "metre" || messageSplit[4] == "metres" || messageSplit[4] == "m" {

			// Initiates the Millimeter to Meter Conversion
			MillimeterToMeter(millimeterNum, s, m)
		} else if messageSplit[4] == "foot" || messageSplit[4] == "feet" || messageSplit[4] == "\"" || messageSplit[4] == "''" || messageSplit[3] == "ft" {

			// Initiates the Millimeter to Foot Conversion
			MillimeterToFoot(millimeterNum, s, m)
		} else if messageSplit[4] == "kilometer" || messageSplit[4] == "kilometers" ||
			messageSplit[4] == "kilometre" || messageSplit[4] == "kilometres" ||
			messageSplit[4] == "km" || messageSplit[4] == "kms" {

			// Initiates the Millimeter to Kilometer Conversion
			MillimeterToKilometer(millimeterNum, s, m)
		} else if messageSplit[4] == "mile" || messageSplit[4] == "miles" || messageSplit[4] == "mi" {

			// Initiates the Millimeter to Mile Conversion
			MillimeterToMile(millimeterNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

///////////////////

// Takes in an Inch number and converts to Centimeter
func InchToCentimeter(inchNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		inchString       string
	)

	// Inch to Centimeter formula
	centimeterNum := inchNum * 2.54

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// Checks if it's 1 inch for grammatically correct output purposes
	if inchNum > 0 && inchNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inch is "+centimeterString+"cm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inches is "+centimeterString+"cm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in an Inch number and converts to Millimetre
func InchToMillimeter(inchNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		millimeterString string
		inchString       string
	)

	// Inch to Millimetre formula
	millimeterNum := inchNum * 25.4

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// Checks if it's 1 inch for grammatically correct output purposes
	if inchNum > 0 && inchNum < 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inch is "+millimeterString+"mm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inches is "+millimeterString+"mm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Inch number and converts to Meter
func InchToMeter(inchNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		inchString  string
		meterString string
	)

	// Inch to Meter formula
	meterNum := inchNum / 1000

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// Checks if it's 1 inch for grammatically correct output purposes
	if inchNum > 0 && inchNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inch is "+meterString+"m")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inches is "+meterString+"m")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Inch number and converts to Foot
func InchToFoot(inchNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		inchString string
		feetString string
	)

	// Inch to Foot formula
	feetNum := inchNum / 12

	// If the result is a whole number it returns that, else the float
	if isFloatInt(feetNum) == true {

		feetNumInt := int(feetNum)

		feetString = strconv.Itoa(feetNumInt)
	} else {

		feetString = strconv.FormatFloat(feetNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// Checks if it's 1 foot and 1 inch for grammatically correct output purposes
	if feetNum > 0 && feetNum <= 1 && (inchNum > 1 || inchNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inches is "+feetString+" foot")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if feetNum > 0 && feetNum <= 1 && (inchNum <= 1 && inchNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inch is "+feetString+" foot")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (feetNum > 1 || feetNum < 0) && (inchNum > 1 || inchNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inches is "+feetString+" feet")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (feetNum > 1 || feetNum < 0) && (inchNum <= 1 && inchNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inch is "+feetString+" feet")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Inch number and converts to Kilometer
func InchToKilometer(inchNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		inchString      string
		kilometerString string
	)

	// Inch to Kilometer formula
	kilometerNum := inchNum * 0.0000254

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', -1, 64)
	}

	// Checks if it's 1 Inch for grammatically correct output purposes
	if inchNum > 0 && inchNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inch is "+kilometerString+"km")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inches is "+kilometerString+"km")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Inch number and converts to Mile
func InchToMile(inchNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		inchString string
		mileString string
	)

	// Inch to mile formula
	mileNum := inchNum / 63360

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', -1, 64)
	}

	// Checks if it's 1 mile and 1 inch for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 && (inchNum > 1 || inchNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inches is "+mileString+" mile")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if mileNum > 0 && mileNum <= 1 && (inchNum <= 1 && inchNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inch is "+mileString+" mile")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (mileNum > 1 || mileNum < 0) && (inchNum > 1 || inchNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inches is "+mileString+" miles")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (mileNum > 1 || mileNum < 0) && (inchNum <= 1 && inchNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, inchString+" inch is "+mileString+" miles")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Inch method
func InchHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper inch number
	inchString := messageSplit[1]

	// Parses the inch number correctly
	if strings.Contains(inchString, "inches") {

		inchString = strings.Replace(inchString, "inches", "", 1)

	} else if strings.Contains(inchString, "inch") {

		inchString = strings.Replace(inchString, "inch", "", 1)

	} else if strings.Contains(inchString, "'") {

		inchString = strings.Replace(inchString, "'", "", 1)
	} else if strings.Contains(inchString, "ins") {

		inchString = strings.Replace(inchString, "ins", "", 1)
	} else if strings.Contains(inchString, "in") {

		inchString = strings.Replace(inchString, "in", "", 1)
	}

	// Converts inch number to a float64 and prints error if it cannot
	inchNum, err := strconv.ParseFloat(inchString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire inch blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "cm" || messageSplit[3] == "cms" || messageSplit[3] == "centimetre" || messageSplit[3] == "centimetres" ||
				messageSplit[3] == "centimeter" || messageSplit[3] == "centimeters" {

				// Initiates the Inch to Centimeter Conversion
				InchToCentimeter(inchNum, s, m)
			} else if messageSplit[3] == "mm" || messageSplit[3] == "mms" || messageSplit[3] == "millimetre" || messageSplit[3] == "millimetres" ||
				messageSplit[3] == "millimeter" || messageSplit[3] == "millimeters" {

				// Initiates the Inch to Millimeter Conversion
				InchToMillimeter(inchNum, s, m)
			} else if messageSplit[3] == "meter" || messageSplit[3] == "meters" ||
				messageSplit[3] == "metre" || messageSplit[3] == "metres" || messageSplit[3] == "m" {

				// Initiates the Inch to Meter Conversion
				InchToMeter(inchNum, s, m)
			} else if messageSplit[3] == "foot" || messageSplit[3] == "feet" || messageSplit[3] == "\"" || messageSplit[3] == "''" || messageSplit[3] == "ft" {

				// Initiates the Inch to Foot Conversion
				InchToFoot(inchNum, s, m)
			} else if messageSplit[3] == "kilometer" || messageSplit[3] == "kilometers" ||
				messageSplit[3] == "kilometre" || messageSplit[3] == "kilometres" ||
				messageSplit[3] == "km" || messageSplit[3] == "kms" {

				// Initiates the Inch to Kilometer Conversion
				InchToKilometer(inchNum, s, m)
			} else if messageSplit[3] == "mile" || messageSplit[3] == "miles" || messageSplit[3] == "mi" {

				// Initiates the Inch to Mile Conversion
				InchToMile(inchNum, s, m)
			}

		} else if messageSplit[4] == "cm" || messageSplit[4] == "cms" || messageSplit[4] == "centimetre" || messageSplit[4] == "centimetres" ||
			messageSplit[4] == "centimeter" || messageSplit[4] == "centimeters" {

			// Initiates the Foot to Centimeter Conversion
			InchToCentimeter(inchNum, s, m)
		} else if messageSplit[4] == "mm" || messageSplit[4] == "mms" || messageSplit[4] == "millimetre" || messageSplit[4] == "millimetres" ||
			messageSplit[4] == "millimeter" || messageSplit[4] == "millimeters" {

			// Initiates the Inch to Millimeter Conversion
			InchToMillimeter(inchNum, s, m)
		} else if messageSplit[4] == "meter" || messageSplit[4] == "meters" ||
			messageSplit[4] == "metre" || messageSplit[4] == "metres" || messageSplit[4] == "m" {

			// Initiates the Inch to Meter Conversion
			InchToMeter(inchNum, s, m)
		} else if messageSplit[4] == "foot" || messageSplit[4] == "feet" || messageSplit[4] == "\"" || messageSplit[4] == "''" || messageSplit[3] == "ft" {

			// Initiates the Inch to Foot Conversion
			InchToFoot(inchNum, s, m)
		} else if messageSplit[4] == "kilometer" || messageSplit[4] == "kilometers" ||
			messageSplit[4] == "kilometre" || messageSplit[4] == "kilometres" ||
			messageSplit[4] == "km" || messageSplit[4] == "kms" {

			// Initiates the Inch to Kilometer Conversion
			InchToKilometer(inchNum, s, m)
		} else if messageSplit[4] == "mile" || messageSplit[4] == "miles" || messageSplit[4] == "mi" {

			// Initiates the Inch to Mile Conversion
			InchToMile(inchNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

///////////////////

// Takes in an Foot number and converts to Centimeter
func FootToCentimeter(footNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		footString       string
	)

	// Foot to Centimeter formula
	centimeterNum := footNum * 30.48

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(footNum) == true {

		footNumInt := int(footNum)

		footString = strconv.Itoa(footNumInt)
	} else {

		footString = strconv.FormatFloat(footNum, 'f', 2, 64)
	}

	// Checks if it's 1 foot for grammatically correct output purposes
	if footNum > 0 && footNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" foot is "+centimeterString+"cm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" feet is "+centimeterString+"cm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in an Foot number and converts to Millimetre
func FootToMillimeter(footNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		millimeterString string
		footString       string
	)

	// Foot to Millimetre formula
	millimeterNum := footNum * 304.8

	// If the result is a whole number it returns that, else the float
	if isFloatInt(footNum) == true {

		footNumInt := int(footNum)

		footString = strconv.Itoa(footNumInt)
	} else {

		footString = strconv.FormatFloat(footNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// Checks if it's 1 foot for grammatically correct output purposes
	if footNum > 0 && footNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" foot is "+millimeterString+"mm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" feet is "+millimeterString+"mm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Foot number and converts to Meter
func FootToMeter(footNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		footString  string
		meterString string
	)

	// Foot to Meter formula
	meterNum := footNum * 0.3048

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(footNum) == true {

		footNumInt := int(footNum)

		footString = strconv.Itoa(footNumInt)
	} else {

		footString = strconv.FormatFloat(footNum, 'f', 2, 64)
	}

	// Checks if it's 1 foot for grammatically correct output purposes
	if footNum > 0 && footNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" foot is "+meterString+"m")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" feet is "+meterString+"m")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Foot number and converts to Inch
func FootToInch(footNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		inchString string
		footString string
	)

	// Foot to Inch formula
	inchNum := footNum * 12

	// If the result is a whole number it returns that, else the float
	if isFloatInt(footNum) == true {

		footNumInt := int(footNum)

		footString = strconv.Itoa(footNumInt)
	} else {

		footString = strconv.FormatFloat(footNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// Checks if it's 1 foot and 1 inch for grammatically correct output purposes
	if footNum > 0 && footNum <= 1 && (inchNum > 1 || inchNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" foot is "+inchString+" inches")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if footNum > 0 && footNum <= 1 && (inchNum <= 1 && inchNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" foot is "+inchString+" inch")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (footNum > 1 || footNum < 0) && (inchNum > 1 || inchNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" feet is "+inchString+" inches")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (footNum > 1 || footNum < 0) && (inchNum <= 1 && inchNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" feet is "+inchString+" inches")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Foot number and converts to Kilometer
func FootToKilometer(footNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		footString      string
		kilometerString string
	)

	// Foot to Kilometer formula
	kilometerNum := footNum * 0.0003048

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(footNum) == true {

		footNumInt := int(footNum)

		footString = strconv.Itoa(footNumInt)
	} else {

		footString = strconv.FormatFloat(footNum, 'f', -1, 64)
	}

	// Checks if it's 1 foot for grammatically correct output purposes
	if footNum > 0 && footNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" foot is "+kilometerString+"km")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" feet is "+kilometerString+"km")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Foot number and converts to Mile
func FootToMile(footNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		footString string
		mileString string
	)

	// Inch to mile formula
	mileNum := footNum / 63360

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(footNum) == true {

		footNumInt := int(footNum)

		footString = strconv.Itoa(footNumInt)
	} else {

		footString = strconv.FormatFloat(footNum, 'f', -1, 64)
	}

	// Checks if it's 1 mile and 1 foot for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 && (footNum > 1 || footNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" feet is "+mileString+" mile")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if mileNum > 0 && mileNum <= 1 && (footNum <= 1 && footNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" foot is "+mileString+" mile")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (mileNum > 1 || mileNum < 0) && (footNum > 1 || footNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" feet is "+mileString+" miles")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (mileNum > 1 || mileNum < 0) && (footNum <= 1 && footNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, footString+" foot is "+mileString+" miles")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Foot method
func FootHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper foot number
	footString := messageSplit[1]

	// Parses the foot number correctly
	if strings.Contains(footString, "foot") {

		footString = strings.Replace(footString, "foot", "", 1)

	} else if strings.Contains(footString, "ft") {

		footString = strings.Replace(footString, "ft", "", 1)

	} else if strings.Contains(footString, "\"") && strings.Contains(footString, "'") == false {

		footString = strings.Replace(footString, "\"", "", 1)
	} else if strings.Contains(footString, "feet") {

		footString = strings.Replace(footString, "feet", "", 1)
	} else if strings.Contains(footString, "''") && strings.Count(footString, "'") < 3 {

		footString = strings.Replace(footString, "''", "", 1)
	}

	// Converts foot number to a float64 and prints error if it cannot
	footNum, err := strconv.ParseFloat(footString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire foot blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "cm" || messageSplit[3] == "cms" || messageSplit[3] == "centimetre" || messageSplit[3] == "centimetres" ||
				messageSplit[3] == "centimeter" || messageSplit[3] == "centimeters" {

				// Initiates the Foot to Centimeter Conversion
				FootToCentimeter(footNum, s, m)
			} else if messageSplit[3] == "mm" || messageSplit[3] == "mms" || messageSplit[3] == "millimetre" || messageSplit[3] == "millimetres" ||
				messageSplit[3] == "millimeter" || messageSplit[3] == "millimeters" {

				// Initiates the Foot to Millimeter Conversion
				FootToMillimeter(footNum, s, m)
			} else if messageSplit[3] == "meter" || messageSplit[3] == "meters" ||
				messageSplit[3] == "metre" || messageSplit[3] == "metres" || messageSplit[3] == "m" {

				// Initiates the Foot to Meter Conversion
				FootToMeter(footNum, s, m)
			} else if messageSplit[3] == "inch" || messageSplit[3] == "inches" ||
				messageSplit[3] == "'" || messageSplit[3] == "in" || messageSplit[3] == "ins" {

				// Initiates the Foot to Inch Conversion
				FootToInch(footNum, s, m)
			} else if messageSplit[3] == "kilometer" || messageSplit[3] == "kilometers" ||
				messageSplit[3] == "kilometre" || messageSplit[3] == "kilometres" ||
				messageSplit[3] == "km" || messageSplit[3] == "kms" {

				// Initiates the Foot to Kilometer Conversion
				FootToKilometer(footNum, s, m)
			} else if messageSplit[3] == "mile" || messageSplit[3] == "miles" || messageSplit[3] == "mi" {

				// Initiates the Foot to Mile Conversion
				FootToMile(footNum, s, m)
			}

		} else if messageSplit[4] == "cm" || messageSplit[4] == "cms" || messageSplit[4] == "centimetre" || messageSplit[4] == "centimetres" ||
			messageSplit[4] == "centimeter" || messageSplit[4] == "centimeters" {

			// Initiates the Foot to Centimeter Conversion
			FootToCentimeter(footNum, s, m)
		} else if messageSplit[4] == "mm" || messageSplit[4] == "mms" || messageSplit[4] == "millimetre" || messageSplit[4] == "millimetres" ||
			messageSplit[4] == "millimeter" || messageSplit[4] == "millimeters" {

			// Initiates the Foot to Millimeter Conversion
			FootToMillimeter(footNum, s, m)
		} else if messageSplit[4] == "meter" || messageSplit[4] == "meters" ||
			messageSplit[4] == "metre" || messageSplit[4] == "metres" || messageSplit[4] == "m" {

			// Initiates the Foot to Meter Conversion
			FootToMeter(footNum, s, m)
		} else if messageSplit[4] == "inch" || messageSplit[4] == "inches" ||
			messageSplit[4] == "'" || messageSplit[4] == "in" || messageSplit[4] == "ins" {

			// Initiates the Foot to Inch Conversion
			FootToInch(footNum, s, m)
		} else if messageSplit[4] == "kilometer" || messageSplit[4] == "kilometers" ||
			messageSplit[4] == "kilometre" || messageSplit[4] == "kilometres" ||
			messageSplit[4] == "km" || messageSplit[4] == "kms" {

			// Initiates the Foot to Kilometer Conversion
			FootToKilometer(footNum, s, m)
		} else if messageSplit[4] == "mile" || messageSplit[4] == "miles" || messageSplit[4] == "mi" {

			// Initiates the Foot to Mile Conversion
			FootToMile(footNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

///////////////////

// Takes in an Meter number and converts to Centimeter
func MeterToCentimeter(meterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		meterString      string
	)

	// Meter to Centimeter formula
	centimeterNum := meterNum * 100

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, meterString+"m is "+centimeterString+"cm")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in an Meter number and converts to Millimetre
func MeterToMillimeter(meterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		millimeterString string
		meterString      string
	)

	// Meter to Millimetre formula
	millimeterNum := meterNum * 1000

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, meterString+"m is "+millimeterString+"mm")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a Meter number and converts to Foot
func MeterToFoot(meterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		footString  string
		meterString string
	)

	// Meter to Foot formula
	footNum := meterNum / 0.3048

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(footNum) == true {

		footNumInt := int(footNum)

		footString = strconv.Itoa(footNumInt)
	} else {

		footString = strconv.FormatFloat(footNum, 'f', 2, 64)
	}

	// Checks if it's 1 foot for grammatically correct output purposes
	if footNum > 0 && footNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, meterString+"m is "+footString+" foot")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, meterString+"m is "+footString+" feet")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Meter number and converts to Inch
func MeterToInch(meterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		meterString string
		inchString  string
	)

	// Meter to Inch formula
	inchNum := meterNum / 0.0254

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		footNumInt := int(meterNum)

		inchString = strconv.Itoa(footNumInt)
	} else {

		inchString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		meterNumInt := int(inchNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// Checks if it's 1 inch for grammatically correct output purposes
	if inchNum > 0 && inchNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, meterString+"m is "+inchString+" inch")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, meterString+"m is "+inchString+" inches")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Meter number and converts to Kilometer
func MeterToKilometer(meterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		meterString     string
		kilometerString string
	)

	// Meter to Kilometer formula
	kilometerNum := meterNum / 1000

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', -1, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, meterString+"m is "+kilometerString+"km")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a Meter number and converts to Mile
func MeterToMile(meterNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		meterString string
		mileString  string
	)

	// Meter to Mile formula
	mileNum := meterNum / 63360

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', -1, 64)
	}

	// Checks if it's 1 mile for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, meterString+"m is "+mileString+" mile")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, meterString+"m is "+mileString+" miles")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Meter method
func MeterHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper foot number
	meterString := messageSplit[1]

	// Parses the meter number correctly
	if strings.Contains(meterString, "meters") {

		meterString = strings.Replace(meterString, "meters", "", 1)

	} else if strings.Contains(meterString, "metres") {

		meterString = strings.Replace(meterString, "metres", "", 1)
	} else if strings.Contains(meterString, "metre") {

		meterString = strings.Replace(meterString, "metre", "", 1)
	} else if strings.Contains(meterString, "meter") {

		meterString = strings.Replace(meterString, "meter", "", 1)
	} else if strings.Contains(meterString, "m") {

		meterString = strings.Replace(meterString, "m", "", 1)
	}

	// Converts meter number to a float64 and prints error if it cannot
	meterNum, err := strconv.ParseFloat(meterString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire meter blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "cm" || messageSplit[3] == "cms" || messageSplit[3] == "centimetre" || messageSplit[3] == "centimetres" ||
				messageSplit[3] == "centimeter" || messageSplit[3] == "centimeters" {

				// Initiates the Meter to Centimeter Conversion
				MeterToCentimeter(meterNum, s, m)
			} else if messageSplit[3] == "mm" || messageSplit[3] == "mms" || messageSplit[3] == "millimetre" || messageSplit[3] == "millimetres" ||
				messageSplit[3] == "millimeter" || messageSplit[3] == "millimeters" {

				// Initiates the Meter to Millimeter Conversion
				MeterToMillimeter(meterNum, s, m)
			} else if messageSplit[3] == "ft" || messageSplit[3] == "feet" ||
				messageSplit[3] == "foot" || messageSplit[3] == "\"" {

				// Initiates the Meter to Foot Conversion
				MeterToFoot(meterNum, s, m)
			} else if messageSplit[3] == "inch" || messageSplit[3] == "inches" ||
				messageSplit[3] == "'" || messageSplit[3] == "in" || messageSplit[3] == "ins" {

				// Initiates the Meter to Inch Conversion
				MeterToInch(meterNum, s, m)
			} else if messageSplit[3] == "kilometer" || messageSplit[3] == "kilometers" ||
				messageSplit[3] == "kilometre" || messageSplit[3] == "kilometres" ||
				messageSplit[3] == "km" || messageSplit[3] == "kms" {

				// Initiates the Meter to Kilometer Conversion
				MeterToKilometer(meterNum, s, m)
			} else if messageSplit[3] == "mile" || messageSplit[3] == "miles" || messageSplit[3] == "mi" {

				// Initiates the Meter to Mile Conversion
				MeterToMile(meterNum, s, m)
			}

		} else if messageSplit[4] == "cm" || messageSplit[4] == "cms" || messageSplit[4] == "centimetre" || messageSplit[4] == "centimetres" ||
			messageSplit[4] == "centimeter" || messageSplit[4] == "centimeters" {

			// Initiates the Meter to Centimeter Conversion
			MeterToCentimeter(meterNum, s, m)
		} else if messageSplit[4] == "mm" || messageSplit[4] == "mms" || messageSplit[4] == "millimetre" || messageSplit[4] == "millimetres" ||
			messageSplit[4] == "millimeter" || messageSplit[4] == "millimeters" {

			// Initiates the Meter to Millimeter Conversion
			MeterToMillimeter(meterNum, s, m)
		} else if messageSplit[4] == "ft" || messageSplit[4] == "feet" ||
			messageSplit[4] == "foot" || messageSplit[4] == "\"" {

			// Initiates the Meter to Foot Conversion
			MeterToFoot(meterNum, s, m)
		} else if messageSplit[4] == "inch" || messageSplit[4] == "inches" ||
			messageSplit[4] == "'" || messageSplit[4] == "in" || messageSplit[4] == "ins" {

			// Initiates the Meter to Inch Conversion
			MeterToInch(meterNum, s, m)
		} else if messageSplit[4] == "kilometer" || messageSplit[4] == "kilometers" ||
			messageSplit[4] == "kilometre" || messageSplit[4] == "kilometres" ||
			messageSplit[4] == "km" || messageSplit[4] == "kms" {

			// Initiates the Meter to Kilometer Conversion
			MeterToKilometer(meterNum, s, m)
		} else if messageSplit[4] == "mile" || messageSplit[4] == "miles" || messageSplit[4] == "mi" {

			// Initiates the Meter to Mile Conversion
			MeterToMile(meterNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

///////////////////

// Takes in an Mile number and converts to Centimeter
func MileToCentimeter(mileNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		mileString       string
	)

	// Mile to Centimeter formula
	centimeterNum := mileNum * 160934.4

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', 2, 64)
	}

	// Checks if it's 1 mile for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" mile is "+centimeterString+"cm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" miles is "+centimeterString+"cm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in an Mile number and converts to Millimetre
func MileToMillimeter(mileNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		millimeterString string
		mileString       string
	)

	// Mile to Millimetre formula
	millimeterNum := mileNum * 1609344

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// Checks if it's 1 mile for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" mile is "+millimeterString+"mm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" miles is "+millimeterString+"mm")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Mile number and converts to Meter
func MileToMeter(mileNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		mileString  string
		meterString string
	)

	// Mile to Meter formula
	meterNum := mileNum * 1609.344

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', 2, 64)
	}

	// Checks if it's 1 mile for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" mile is "+meterString+"m")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" miles is "+meterString+"m")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Mile number and converts to Inch
func MileToInch(mileNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		inchString string
		mileString string
	)

	// Mile to Inch formula
	inchNum := mileNum * 63360

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// Checks if it's 1 mile and 1 inch for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 && (inchNum > 1 || inchNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" mile is "+inchString+" inches")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if mileNum > 0 && mileNum <= 1 && inchNum <= 1 && inchNum > 0 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" mile is "+inchString+" inch")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (mileNum > 1 || mileNum < 0) && (inchNum > 1 || inchNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" miles is "+inchString+" inches")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (mileNum > 1 || mileNum < 0) && (inchNum <= 1 && inchNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" miles is "+inchString+" inch")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Mile number and converts to Kilometer
func MileToKilometer(mileNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		mileString      string
		kilometerString string
	)

	// Mile to Kilometer formula
	kilometerNum := mileNum * 1.609344

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', -1, 64)
	}

	// Checks if it's 1 mile for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" mile is "+kilometerString+"km")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" miles is "+kilometerString+"km")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Mile number and converts to Foot
func MileToFoot(mileNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		footString string
		mileString string
	)

	// Mile to Foot formula
	footNum := mileNum * 5280

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(footNum) == true {

		footNumInt := int(footNum)

		footString = strconv.Itoa(footNumInt)
	} else {

		footString = strconv.FormatFloat(footNum, 'f', -1, 64)
	}

	// Checks if it's 1 mile and 1 foot for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 && (footNum > 1 || footNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" mile is "+footString+" feet")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if mileNum > 0 && mileNum <= 1 && (footNum < 1 && footNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" mile is "+footString+" foot")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (mileNum > 1 || mileNum < 0) && (footNum > 1 || footNum < 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" miles is "+footString+" feet")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else if (mileNum > 1 || mileNum < 0) && (footNum <= 1 && footNum > 0) {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, mileString+" miles is "+footString+" foot")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Mile method
func MileHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper mile number
	mileString := messageSplit[1]

	// Parses the mile number correctly
	if strings.Contains(mileString, "miles") {

		mileString = strings.Replace(mileString, "miles", "", 1)

	} else if strings.Contains(mileString, "mile") {

		mileString = strings.Replace(mileString, "mile", "", 1)

	} else if strings.Contains(mileString, "mi") {

		mileString = strings.Replace(mileString, "mi", "", 1)
	}

	// Converts mile number to a float64 and prints error if it cannot
	mileNum, err := strconv.ParseFloat(mileString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire mile blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "cm" || messageSplit[3] == "cms" || messageSplit[3] == "centimetre" || messageSplit[3] == "centimetres" ||
				messageSplit[3] == "centimeter" || messageSplit[3] == "centimeters" {

				// Initiates the Mile to Centimeter Conversion
				MileToCentimeter(mileNum, s, m)
			} else if messageSplit[3] == "mm" || messageSplit[3] == "mms" || messageSplit[3] == "millimetre" || messageSplit[3] == "millimetres" ||
				messageSplit[3] == "millimeter" || messageSplit[3] == "millimeters" {

				// Initiates the Mile to Millimeter Conversion
				MileToMillimeter(mileNum, s, m)
			} else if messageSplit[3] == "meter" || messageSplit[3] == "meters" ||
				messageSplit[3] == "metre" || messageSplit[3] == "metres" || messageSplit[3] == "m" {

				// Initiates the Mile to Meter Conversion
				MileToMeter(mileNum, s, m)
			} else if messageSplit[3] == "inch" || messageSplit[3] == "inches" ||
				messageSplit[3] == "'" || messageSplit[3] == "in" || messageSplit[3] == "ins" {

				// Initiates the Mile to Inch Conversion
				MileToInch(mileNum, s, m)
			} else if messageSplit[3] == "kilometer" || messageSplit[3] == "kilometers" ||
				messageSplit[3] == "kilometre" || messageSplit[3] == "kilometres" ||
				messageSplit[3] == "km" || messageSplit[3] == "kms" {

				// Initiates the Mile to Kilometer Conversion
				MileToKilometer(mileNum, s, m)
			} else if messageSplit[3] == "ft" || messageSplit[3] == "feet" || messageSplit[3] == "foot" || messageSplit[3] == "\"" || messageSplit[3] == "''" {

				// Initiates the Mile to Foot Conversion
				MileToFoot(mileNum, s, m)
			}

		} else if messageSplit[4] == "cm" || messageSplit[4] == "cms" || messageSplit[4] == "centimetre" || messageSplit[4] == "centimetres" ||
			messageSplit[4] == "centimeter" || messageSplit[4] == "centimeters" {

			// Initiates the Mile to Centimeter Conversion
			MileToCentimeter(mileNum, s, m)
		} else if messageSplit[4] == "mm" || messageSplit[4] == "mms" || messageSplit[4] == "millimetre" || messageSplit[4] == "millimetres" ||
			messageSplit[4] == "millimeter" || messageSplit[4] == "millimeters" {

			// Initiates the Mile to Millimeter Conversion
			MileToMillimeter(mileNum, s, m)
		} else if messageSplit[4] == "meter" || messageSplit[4] == "meters" ||
			messageSplit[4] == "metre" || messageSplit[4] == "metres" || messageSplit[4] == "m" {

			// Initiates the Mile to Meter Conversion
			MileToMeter(mileNum, s, m)
		} else if messageSplit[4] == "inch" || messageSplit[4] == "inches" ||
			messageSplit[4] == "'" || messageSplit[4] == "in" || messageSplit[4] == "ins" {

			// Initiates the Mile to Inch Conversion
			MileToInch(mileNum, s, m)
		} else if messageSplit[4] == "kilometer" || messageSplit[4] == "kilometers" ||
			messageSplit[4] == "kilometre" || messageSplit[4] == "kilometres" ||
			messageSplit[4] == "km" || messageSplit[4] == "kms" {

			// Initiates the Mile to Kilometer Conversion
			MileToKilometer(mileNum, s, m)
		} else if messageSplit[4] == "foot" || messageSplit[4] == "feet" || messageSplit[4] == "ft" || messageSplit[4] == "\"" || messageSplit[4] == "''" {

			// Initiates the Mile to Foot Conversion
			MileToFoot(mileNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

///////////////////

// Takes in a Kilometer number and converts to Millimeter
func KilometerToMillimeter(kilometerNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		kilometerString  string
		millimeterString string
	)

	// Kilometer to Millimeter formula
	millimeterNum := kilometerNum * 1000000

	// If the result is a whole number it returns that, else the float
	if isFloatInt(millimeterNum) == true {

		millimeterNumInt := int(millimeterNum)

		millimeterString = strconv.Itoa(millimeterNumInt)
	} else {

		millimeterString = strconv.FormatFloat(millimeterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', 2, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, kilometerString+"km is "+millimeterString+"mm")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a Kilometer number and converts to Inch
func KilometerToInch(kilometerNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		kilometerString string
		inchString      string
	)

	// Kilometer to Inch formula
	inchNum := kilometerNum / 0.0000254

	// If the result is a whole number it returns that, else the float
	if isFloatInt(inchNum) == true {

		inchNumInt := int(inchNum)

		inchString = strconv.Itoa(inchNumInt)
	} else {

		inchString = strconv.FormatFloat(inchNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', 2, 64)
	}

	// Checks if it's 1 inch for grammatically correct output purposes
	if inchNum > 0 && inchNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, kilometerString+"km is "+inchString+" inch")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, kilometerString+"km is "+inchString+" inches")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Kilometer number and converts to Meter
func KilometerToMeter(kilometerNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		kilometerString string
		meterString     string
	)

	// Kilometer to Meter formula
	meterNum := kilometerNum * 1000

	// If the result is a whole number it returns that, else the float
	if isFloatInt(meterNum) == true {

		meterNumInt := int(meterNum)

		meterString = strconv.Itoa(meterNumInt)
	} else {

		meterString = strconv.FormatFloat(meterNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', 2, 64)
	}

	// Checks if it's 1 meter for grammatically correct output purposes
	if meterNum > 0 && meterNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, kilometerString+"km is "+meterString+" meter")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, kilometerString+"km is "+meterString+" meters")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Kilometer number and converts to Foot
func KilometerToFoot(kilometerNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		kilometerString string
		feetString      string
	)

	// Kilometer to Foot formula
	feetNum := kilometerNum / 0.0003048

	// If the result is a whole number it returns that, else the float
	if isFloatInt(feetNum) == true {

		feetNumInt := int(feetNum)

		feetString = strconv.Itoa(feetNumInt)
	} else {

		feetString = strconv.FormatFloat(feetNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', 2, 64)
	}

	// Checks if it's 1 foot for grammatically correct output purposes
	if feetNum > 0 && feetNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, kilometerString+"km is "+feetString+" foot")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, kilometerString+"km is "+feetString+" feet")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Takes in a Kilometer number and converts to Centimeter
func KilometerToCentimeter(kilometerNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		centimeterString string
		kilometerString  string
	)

	// Centimeter to Kilometer formula
	centimeterNum := kilometerNum * 100000

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', -1, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(centimeterNum) == true {

		centimeterNumInt := int(centimeterNum)

		centimeterString = strconv.Itoa(centimeterNumInt)
	} else {

		centimeterString = strconv.FormatFloat(centimeterNum, 'f', -1, 64)
	}

	// Prints result
	_, err := s.ChannelMessageSend(m.ChannelID, centimeterString+"km is "+kilometerString+"cm")
	if err != nil {

		fmt.Println("Error: ", err)
	}
}

// Takes in a Kilometer number and converts to Mile
func KilometerToMile(kilometerNum float64, s discordgo.Session, m discordgo.MessageCreate) {

	var (
		kilometerString string
		mileString      string
	)

	// Centimeter to mile formula
	mileNum := kilometerNum / 1.609344

	// If the result is a whole number it returns that, else the float
	if isFloatInt(mileNum) == true {

		mileNumInt := int(mileNum)

		mileString = strconv.Itoa(mileNumInt)
	} else {

		mileString = strconv.FormatFloat(mileNum, 'f', 2, 64)
	}

	// If the result is a whole number it returns that, else the float
	if isFloatInt(kilometerNum) == true {

		kilometerNumInt := int(kilometerNum)

		kilometerString = strconv.Itoa(kilometerNumInt)
	} else {

		kilometerString = strconv.FormatFloat(kilometerNum, 'f', 2, 64)
	}

	// Checks if it's 1 mile for grammatically correct output purposes
	if mileNum > 0 && mileNum <= 1 {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, kilometerString+"km is "+mileString+" mile")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	} else {
		// Prints result
		_, err := s.ChannelMessageSend(m.ChannelID, kilometerString+"km is "+mileString+" miles")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}
}

// Kilometer method
func KilometerHandler(messageSplit []string, s discordgo.Session, m discordgo.MessageCreate) {

	// Create bool variable that checks if wrong string is contained in the string, stopping command from executing if true
	var badString bool

	// Takes the proper kilometer number
	kilometerString := messageSplit[1]

	// Parses the kilometer number correctly
	if strings.Contains(kilometerString, "kilometers") {

		kilometerString = strings.Replace(kilometerString, "kilometers", "", 1)

	} else if strings.Contains(kilometerString, "kilometres") {

		kilometerString = strings.Replace(kilometerString, "kilometres", "", 1)

	} else if strings.Contains(kilometerString, "kms") {

		kilometerString = strings.Replace(kilometerString, "kms", "", 1)
	} else if strings.Contains(kilometerString, "km") {

		kilometerString = strings.Replace(kilometerString, "km", "", 1)
	} else if strings.Contains(kilometerString, "kilometre") {

		kilometerString = strings.Replace(kilometerString, "kilometre", "", 1)
	} else if strings.Contains(kilometerString, "kilometer") {

		kilometerString = strings.Replace(kilometerString, "kilometer", "", 1)
	}

	// Converts kilometer number to a float64 and prints error if it cannot
	centimeterNum, err := strconv.ParseFloat(kilometerString, 64)
	if err != nil {

		badString = true

		// Prints error message
		_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert from that measurement.")
		if err != nil {

			fmt.Println("Error: ", err)
		}
	}

	if badString == false {
		// Entire kilometer blurb for every possible messageSplit length
		if len(messageSplit) == 4 {

			if messageSplit[3] == "mm" || messageSplit[3] == "mms" || messageSplit[3] == "millimetre" || messageSplit[3] == "millimetres" ||
				messageSplit[3] == "millimeter" || messageSplit[3] == "millimeters" {

				// Initiates the Kilometer to Millimeter Conversion
				KilometerToMillimeter(centimeterNum, s, m)
			} else if messageSplit[3] == "inch" || messageSplit[3] == "inches" ||
				messageSplit[3] == "'" || messageSplit[3] == "in" || messageSplit[3] == "ins" {

				// Initiates the Kilometer to Inch Conversion
				KilometerToInch(centimeterNum, s, m)
			} else if messageSplit[3] == "meter" || messageSplit[3] == "meters" ||
				messageSplit[3] == "metre" || messageSplit[3] == "metres" || messageSplit[3] == "m" {

				// Initiates the Kilometer to Meter Conversion
				KilometerToMeter(centimeterNum, s, m)
			} else if messageSplit[3] == "foot" || messageSplit[3] == "feet" || messageSplit[3] == "\"" || messageSplit[3] == "''" || messageSplit[3] == "ft" {

				// Initiates the Kilometer to Foot Conversion
				KilometerToFoot(centimeterNum, s, m)
			} else if messageSplit[3] == "centimeter" || messageSplit[3] == "centimeters" ||
				messageSplit[3] == "cm" || messageSplit[3] == "cms" ||
				messageSplit[3] == "centimetre" || messageSplit[3] == "centimetres" {

				// Initiates the Kilometer to Centimeter Conversion
				KilometerToCentimeter(centimeterNum, s, m)
			} else if messageSplit[3] == "mile" || messageSplit[3] == "miles" || messageSplit[3] == "mi" {

				// Initiates the Kilometer to Mile Conversion
				KilometerToMile(centimeterNum, s, m)
			}

		} else if messageSplit[4] == "mm" || messageSplit[4] == "mms" || messageSplit[4] == "millimetre" || messageSplit[4] == "millimetres" ||
			messageSplit[4] == "millimeter" || messageSplit[4] == "millimeters" {

			// Initiates the Kilometer to Millimeter Conversion
			KilometerToMillimeter(centimeterNum, s, m)
		} else if messageSplit[4] == "inch" || messageSplit[4] == "inches" ||
			messageSplit[4] == "'" || messageSplit[4] == "in" || messageSplit[4] == "ins" {

			// Initiates the Kilometer to Inch Conversion
			KilometerToInch(centimeterNum, s, m)
		} else if messageSplit[4] == "meter" || messageSplit[4] == "meters" ||
			messageSplit[4] == "metre" || messageSplit[4] == "metres" || messageSplit[3] == "m" {

			// Initiates the Kilometer to Meter Conversion
			KilometerToMeter(centimeterNum, s, m)
		} else if messageSplit[4] == "foot" || messageSplit[4] == "feet" || messageSplit[4] == "\"" || messageSplit[4] == "''" || messageSplit[3] == "ft" {

			// Initiates the Kilometer to Foot Conversion
			KilometerToFoot(centimeterNum, s, m)
		} else if messageSplit[4] == "centimeter" || messageSplit[4] == "centimeters" ||
			messageSplit[4] == "cm" || messageSplit[4] == "cms" ||
			messageSplit[4] == "centimetre" || messageSplit[4] == "centimetres" {

			// Initiates the Kilometer to Centimeter Conversion
			KilometerToCentimeter(centimeterNum, s, m)
		} else if messageSplit[4] == "mile" || messageSplit[4] == "miles" || messageSplit[4] == "mi" {

			// Initiates the Kilometer to Mile Conversion
			KilometerToMile(centimeterNum, s, m)
		} else {

			// Prints error message
			_, err := s.ChannelMessageSend(m.ChannelID, "Error: cannot convert that.")
			if err != nil {

				fmt.Println("Error: ", err)
			}
		}
	}
}

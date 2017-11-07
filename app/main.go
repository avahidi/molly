package main

import (
	"fmt"

	"bitbucket.org/vahidi/molly"
	"bitbucket.org/vahidi/molly/at"
)

func main() {

	c := at.NewClass("DlinkFirmwareHeader")
	c.AddAssignment("magic",
		&at.ExtractExpression{
			Offset: &at.NumberExpression{Value: 0},
			Size:   &at.NumberExpression{Value: 4},
		})

	c.AddAssignment("bootnameadr",
		&at.ExtractExpression{
			Offset: &at.NumberExpression{Value: 7},
			Size:   &at.NumberExpression{Value: 1},
		})

	c.AddAssignment("bootname", &at.ExtractExpression{
		Offset: &at.BinaryExpression{
			Left:      &at.VariableExpression{Id: "bootnameadr"},
			Right:     &at.NumberExpression{Value: 12},
			Operation: at.ADD,
		},
		Size: &at.NumberExpression{Value: 4},
	})

	c.AddCondition(&at.BinaryExpression{
		Left:      &at.VariableExpression{Id: "bootname"},
		Right:     &at.NumberExpression{Value: 0x54a3a417},
		Operation: at.EQ,
	})

	c2 := at.NewClass("UImage")
	c2.AddAssignment("magic",
		&at.ExtractExpression{
			Offset: &at.NumberExpression{Value: 0},
			Size:   &at.NumberExpression{Value: 4},
		})

	c2.AddCondition(&at.BinaryExpression{
		Left:      &at.VariableExpression{Id: "magic"},
		Right:     &at.NumberExpression{Value: 0x27051956},
		Operation: at.EQ,
	})

	c2.AddAction("echo", "bla", "blabla")

	db := molly.NewDatabase()
	db.Classes = append(db.Classes, c)
	db.Classes = append(db.Classes, c2)

	err := db.ScanFile("/home/work/tmp/fw/")
	if err != nil {
		fmt.Println("SCAN ERROR: ", err)
	}
}

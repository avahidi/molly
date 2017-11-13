package main

import (
	"fmt"

	"bitbucket.org/vahidi/molly/prim"

	"bitbucket.org/vahidi/molly"
	"bitbucket.org/vahidi/molly/at"
)

func main() {

	/*
			// simplification test
			c0 := at.NewClass("Test1")

			c0.AddAssignment("dummy", &at.BinaryExpression{
				Left:      at.NewNumberExpression(12, 8, prim.UBE),
				Right:     at.NewNumberExpression(16, 8, prim.UBE),
				Operation: prim.ADD,
			})

			c0.AddAssignment("yummy", &at.BinaryExpression{
				Left:      &at.VariableExpression{Id: "doesn't exist"},
				Right:     at.NewNumberExpression(16, 8, prim.UBE),
				Operation: prim.ADD,
			})


		fmt.Println("BEFORE", c0)
		c0.Simplify()
		fmt.Println("AFTER", c0)
	*/
	c := at.NewClass("DlinkFirmwareHeader")
	c.AddAssignment("magic",
		&at.ExtractExpression{
			Offset: at.NewNumberExpression(0, 8, prim.UBE),
			Size:   at.NewNumberExpression(4, 8, prim.UBE),
		})

	c.AddAssignment("bootnameadr",
		&at.ExtractExpression{
			Offset: at.NewNumberExpression(7, 8, prim.UBE),
			Size:   at.NewNumberExpression(1, 8, prim.UBE),
		})

	c.AddAssignment("bootname", &at.ExtractExpression{
		Offset: &at.BinaryExpression{
			Left:      &at.VariableExpression{Id: "bootnameadr"},
			Right:     at.NewNumberExpression(12, 8, prim.UBE),
			Operation: prim.ADD,
		},
		Size: at.NewNumberExpression(4, 8, prim.UBE),
	})

	c.AddCondition(&at.BinaryExpression{
		Left:      &at.VariableExpression{Id: "bootname"},
		Right:     at.NewNumberExpression(0x54a3a417, 4, prim.UBE),
		Operation: prim.EQ,
	})

	c2 := at.NewClass("UImage")
	c2.AddAssignment("magic",
		&at.ExtractExpression{
			Offset: at.NewNumberExpression(0, 8, prim.UBE),
			Size:   at.NewNumberExpression(4, 8, prim.UBE),
		})

	c2.AddAssignment("tragic",
		&at.UnaryExpression{
			Value:     &at.VariableExpression{Id: "filesize$"},
			Operation: prim.NEG,
		})

	c2.AddAssignment("kragic",
		&at.BinaryExpression{
			Left: &at.BinaryExpression{
				Left:      at.NewStringExpression("Zappa"),
				Right:     at.NewStringExpression("Kappa"),
				Operation: prim.ADD,
			},
			Right:     at.NewStringExpression("ZappaKappa"),
			Operation: prim.EQ,
		})

	c2.AddCondition(&at.BinaryExpression{
		Left:      &at.VariableExpression{Id: "magic"},
		Right:     at.NewNumberExpression(0x27051956, 4, prim.UBE),
		Operation: prim.EQ,
	})

	c2.AddAction("echo", "bla", "blabla")

	db := molly.NewDatabase()
	db.Classes = append(db.Classes, c)
	db.Classes = append(db.Classes, c2)

	fmt.Println("DB=", db)
	err := db.ScanFile("/home/work/tmp/fw/")
	if err != nil {
		fmt.Println("SCAN ERROR: ", err)
	}
}

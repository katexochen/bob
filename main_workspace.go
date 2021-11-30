package main

import (
	"github.com/spf13/cobra"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"
)

var cmdWorkspace = &cobra.Command{
	Use:   "workspace",
	Short: "Manage a bob workspace",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

func runInit() {
	bob, err := bob.Bob()
	errz.Fatal(err)

	err = bob.Init()
	errz.Fatal(err)
}
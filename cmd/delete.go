package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tczekajlo/kir/etcd"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a rule",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("You have to give a name of rule to delete")
			return
		}
		etcd := etcd.Client{}
		etcd.New()
		err := etcd.Delete(args[0])
		if err != nil {
			fmt.Println("Cannot delete rule:", err)
			return
		}
		defer etcd.Client.Close()

		fmt.Println("Deleted rules:", etcd.DeleteResponse.Deleted)

	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}

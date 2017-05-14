package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tczekajlo/kir/config"
	"github.com/tczekajlo/kir/etcd"
	"github.com/tczekajlo/kir/pb"
)

func getAllRules(cmd *cobra.Command, args []string) {
	var err error
	var data *pb.RulesList
	var dataRule *pb.Rule
	var limit int64

	// etcd
	etcd := etcd.Client{}
	etcd.New()

	if len(args) == 0 {
		limit = config.EtcdGetLimit
		if showAll, _ := cmd.Flags().GetBool("show-all"); showAll {
			limit = 0
		}

		data, err = etcd.GetAll(limit)
	} else {
		dataRule, err = etcd.Get("rule/" + args[0])
		data = &pb.RulesList{}
		data.Rule = append(data.Rule, dataRule)

	}
	if err != nil {
		fmt.Println("Cannot get rule(s):", err)
		return
	}
	defer etcd.Client.Close()

	//print output
	if len(args) != 0 && cmd.Flag("output").Value.String() != "" {
		printOutput(dataRule, cmd.Flag("output").Value.String())
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Namespace", "Image", "Annotations", "Allowed"})
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator(" ")
	table.SetColumnSeparator(" ")

	for _, rule := range data.Rule {
		var image []string
		rule = ruleFillDefault(rule)

		for _, container := range rule.Containers {
			image = append(image, container.Image)
		}
		table.Append([]string{rule.Name,
			rule.Namespace,
			strings.Join(image, "\n"),
			strings.Join(annotationsToString(rule.Annotations), "\n"),
			strconv.FormatBool(rule.Allowed),
		})
	}
	table.Render()

	if etcd.GetResponse.Count > limit && limit != 0 {
		fmt.Println("Showed results are limited. In order to show all results use --show-all flag.")
	}
}

func printOutput(data *pb.Rule, format string) {
	switch format {
	case "yaml":
		output, err := yaml.Marshal(data)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Print(string(output))
	default:
		fmt.Printf("Format %s is not supported\n", format)
	}
}

func annotationsToString(annotations map[string]string) []string {
	var result []string

	for key, value := range annotations {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}

	return result
}

func ruleFillDefault(data *pb.Rule) *pb.Rule {
	if data.Annotations == nil {
		data.Annotations = make(map[string]string)
		data.Annotations["<none>"] = "<none>"
	}

	return data
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets the rule or all rules",
	Run: func(cmd *cobra.Command, args []string) {
		getAllRules(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(getCmd)

	getCmd.Flags().StringP("output", "o", "", "set the output format (yaml)")
	getCmd.Flags().BoolP("show-all", "a", false, "show all rules")
}

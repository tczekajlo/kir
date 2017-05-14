package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/spf13/cobra"
	"github.com/tczekajlo/kir/etcd"
	"github.com/tczekajlo/kir/pb"
)

type ruleConfig struct {
	Name        string
	Image       []string
	Allowed     bool
	Annotations []string
	Namespace   string
	Reason      string
}

var rule ruleConfig

func setupValidate() error {
	if len(rule.Image) == 0 {
		return fmt.Errorf("List of images is empty. Use --image flag")
	}

	if rule.Name == "" {
		return fmt.Errorf("Rule name is empty. Use --name flag")
	}

	if rule.Namespace == "" {
		return fmt.Errorf("Namespace name is empty. Use --namespace flag")
	}

	return nil
}

func setupRule(data *pb.Rule) (*pb.Rule, error) {
	var err error

	if len(rule.Annotations) != 0 {
		data.Annotations, err = rule.annotationsToMap()
		if err != nil {
			return data, fmt.Errorf("%s", err)
		}
	}

	data.Containers = rule.imageToContainerImage()

	return data, nil
}

func (r *ruleConfig) annotationsToMap() (map[string]string, error) {
	var result = make(map[string]string)
	for _, annotation := range rule.Annotations {
		data := strings.Split(annotation, "=")
		if len(data) < 2 {
			return result, fmt.Errorf("Cannot parse annotation")
		}

		result[data[0]] = data[1]
	}

	return result, nil
}

func (r *ruleConfig) imageToContainerImage() []*pb.Rule_Containers {
	var result []*pb.Rule_Containers
	for _, image := range rule.Image {
		result = append(result, &pb.Rule_Containers{Image: image})
	}

	return result
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new rule",
	Long: `Adds a new rule which will be taken into account during image's review.
For example:

# Allows for run PODs with nginx image in 1.x version within default namespace
kir add --name my_rule --image ^nginx:1.[0-9] --allowed --namespace ^default$

`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var data *pb.Rule
		var fileData []byte

		if cmd.Flag("file").Value.String() == "" {
			err = setupValidate()
			if err != nil {
				fmt.Println(err)
				return
			}

			data, err = setupRule(&pb.Rule{
				Name:      rule.Name,
				Allowed:   rule.Allowed,
				Namespace: rule.Namespace,
				Reason:    rule.Reason,
			})
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			fileData, err = ioutil.ReadFile(cmd.Flag("file").Value.String())
			if err != nil {
				fmt.Printf("err: %v\n", err)
				return
			}

			data = &pb.Rule{}
			err = yaml.Unmarshal(fileData, data)
			if err != nil {
				fmt.Printf("err: %v\n", err)
				return
			}
		}

		// etcd
		etcd := etcd.Client{}
		etcd.New()
		override, _ := cmd.Flags().GetBool("override")
		err = etcd.Add(data, override)
		if err != nil {
			fmt.Println("Cannot add rule")
			return
		}
		defer etcd.Client.Close()

		if etcd.TxnResponse.Succeeded {
			fmt.Printf("Rule \"%s\" added.\n", data.Name)
		} else {
			fmt.Printf("Rule \"%s\" already exists.\n", data.Name)
		}
	},
}

func init() {
	rule = ruleConfig{}

	RootCmd.AddCommand(addCmd)

	addCmd.Flags().StringSliceVar(&rule.Image, "image", []string{}, "container image name (items in a list should be separated by a comma)")
	addCmd.Flags().StringSliceVar(&rule.Annotations, "annotations", []string{}, "list of annotations, e.g. key=value (items in a list should be separated by a comma)")
	addCmd.Flags().StringVar(&rule.Namespace, "namespace", "", "namespace name")
	addCmd.Flags().StringVar(&rule.Name, "name", "", "rule name")
	addCmd.Flags().StringVar(&rule.Reason, "reason", "", "reason why for this rule is blocking image")
	addCmd.Flags().BoolVar(&rule.Allowed, "allowed", false, "action to take if request match to a rule")
	addCmd.Flags().Bool("override", false, "override existing rule")

	addCmd.Flags().StringP("file", "f", "", "add rule based on data from a file")
}

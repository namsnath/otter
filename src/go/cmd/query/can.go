package cmd

import (
	"fmt"

	"github.com/namsnath/otter/action"
	"github.com/namsnath/otter/query"
	"github.com/namsnath/otter/resource"
	"github.com/namsnath/otter/specifier"
	"github.com/namsnath/otter/subject"
	"github.com/spf13/cobra"
)

var canCmd = &cobra.Command{
	Use:   "can subject",
	Short: "Check if a subject can perform an action on a resource with given specifiers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return nil
		}

		subjectStr := args[0]

		subjectType, subjectTypeErr := subject.SubjectTypeFromString(cmd.Flag("of-type").Value.String())
		if subjectTypeErr != nil {
			return subjectTypeErr
		}

		action, actionErr := action.FromString(cmd.Flag("perform").Value.String())
		if actionErr != nil {
			return actionErr
		}

		on := cmd.Flag("on").Value.String()
		resource := resource.Resource{Name: on}

		specifierMap, err := cmd.Flags().GetStringToString("with")
		if err != nil {
			return err
		}

		specifiers := []specifier.Specifier{}
		for k, v := range specifierMap {
			specifiers = append(specifiers, specifier.Specifier{Key: k, Value: v})
		}
		specifierGroup := specifier.SpecifierGroup{Specifiers: specifiers}

		can := query.Can(subject.Subject{Name: subjectStr, Type: subjectType}).Perform(action).On(resource).With(specifierGroup).Query()
		if can.Err != nil {
			return err
		}

		fmt.Println(can.Pretty())

		return nil
	},
}

func init() {
	QueryCmd.AddCommand(canCmd)

	canCmd.Args = cobra.ExactArgs(1)

	canCmd.Flags().String("of-type", string(subject.SubjectTypePrincipal), "The type of subject")
	canCmd.Flags().String("perform", "", "Action to check permission for")
	canCmd.Flags().String("on", "", "Parent resource under which to check permissions")
	canCmd.Flags().StringToString("with", map[string]string{}, "Map of specifiers to check permissions with. Format: key1=value1,key2=value2")
}

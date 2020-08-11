package cmd

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/elastic/elastic-package/internal/promote"
)

func setupPromoteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "promote",
		Short:        "Promote the package",
		Long:         "Use promote command to move packages between stages in package-storage.",
		RunE:         promoteCommandAction,
		SilenceUsage: true,
	}
	return cmd
}

func promoteCommandAction(cmd *cobra.Command, args []string) error {
	sourceStage, destinationStage, err := promptPromotion()
	if err != nil {
		return errors.Wrap(err, "prompt for promotion failed")
	}

	newestOnly, err := promptPromoteNewestOnly()
	if err != nil {
		return errors.Wrap(err, "prompt for promoting newest revisions only failed")
	}

	repository, err := promote.CloneRepository(sourceStage)
	if err != nil {
		return errors.Wrapf(err, "cloning source repository failed (branch: %s)", sourceStage)
	}

	allPackages, err := promote.ListPackages(repository, newestOnly)
	if err != nil {
		return errors.Wrapf(err, "listing packages failed (newestOnly: %t)", newestOnly)
	}

	if len(allPackages) == 0 {
		fmt.Println("No packages available for promotion.")
		return nil
	}

	promotedPackages, err := promptPackages(allPackages)
	if err != nil {
		return errors.Wrap(err, "prompt for package selection failed")
	}

	removedPackages := promote.DetermineRemovedPackages(allPackages, promotedPackages, newestOnly)

	// Copy packages to destination
	newDestinationStage, err := promote.CopyPackages(repository, sourceStage, destinationStage, promotedPackages)
	if err != nil {
		return errors.Wrapf(err, "copying packages failed (source: %s, destination: %s)", sourceStage, destinationStage)
	}

	// Remove packages from source
	newSourceStage, err := promote.RemovePackages(repository, sourceStage, removedPackages)
	if err != nil {
		return errors.Wrapf(err, "removing packages failed (source: %s)", sourceStage)
	}

	// Push changes
	err = promote.PushChanges(repository, newDestinationStage)
	if err != nil {
		return errors.Wrapf(err, "pushing changes failed (stage: %s)", newDestinationStage)
	}

	err = promote.PushChanges(repository, newDestinationStage)
	if err != nil {
		return errors.Wrapf(err, "pushing changes failed (stage: %s)", newSourceStage)
	}

	// Open PRs
	err = promote.OpenPullRequestWithPromotedPackages(newDestinationStage, destinationStage, promotedPackages)
	if err != nil {
		return errors.Wrapf(err, "opening PR with promoted packages failed (head: %s, base: %s)", newDestinationStage, destinationStage)
	}

	err = promote.OpenPullRequestWithRemovedPackages(newDestinationStage, destinationStage, removedPackages)
	if err != nil {
		return errors.Wrapf(err, "opening PR with removed packages failed (head: %s, base: %s)", newDestinationStage, destinationStage)
	}
	return nil
}

func promptPromotion() (string, string, error) {
	promotionPrompt := &survey.Select{
		Message: "Which promotion would you like to run",
		Options: []string{"snapshot - staging", "staging - production", "snapshot - production"},
		Default: "snapshot - staging",
	}

	var promotion string
	err := survey.AskOne(promotionPrompt, &promotion)
	if err != nil {
		return "", "", err
	}

	s := strings.Split(promotion, " - ")
	return s[0], s[1], nil
}

func promptPromoteNewestOnly() (bool, error) {
	newestOnly := true
	prompt := &survey.Confirm{
		Message: "Would you like to promote newest revisions only and remove older ones?",
		Default: true,
	}
	err := survey.AskOne(prompt, &newestOnly)
	if err != nil {
		return false, err
	}
	return newestOnly, nil
}

func promptPackages(allPackages promote.PackageRevisions) (promote.PackageRevisions, error) {
	packagesPrompt := &survey.MultiSelect{
		Message:  "Which packages would you like to promote",
		Options:  allPackages.Strings(),
		PageSize: 100,
	}

	var selectedOptions []string
	err := survey.AskOne(packagesPrompt, &selectedOptions, survey.WithValidator(survey.Required))
	if err != nil {
		return nil, err
	}

	var selected promote.PackageRevisions
	for _, option := range selectedOptions {
		for _, p := range allPackages {
			if p.String() == option {
				selected = append(selected, p)
			}
		}
	}
	return selected, nil
}

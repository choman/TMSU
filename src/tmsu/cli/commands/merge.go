/*
Copyright 2011-2013 Paul Ruane.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package commands

import (
	"fmt"
	"tmsu/cli"
	"tmsu/log"
	"tmsu/storage"
)

type MergeCommand struct {
	verbose bool
}

func (MergeCommand) Name() cli.CommandName {
	return "merge"
}

func (MergeCommand) Synopsis() string {
	return "Merge tags"
}

func (MergeCommand) Description() string {
	return `tmsu merge TAG... DEST
        
Merges TAGs into tag DEST resulting in a single tag of name DEST.`
}

func (MergeCommand) Options() cli.Options {
	return cli.Options{}
}

func (command MergeCommand) Exec(options cli.Options, args []string) error {
	command.verbose = options.HasOption("--verbose")

	if len(args) < 2 {
		return fmt.Errorf("too few arguments.")
	}

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	destTagName := args[len(args)-1]
	destTag, err := store.TagByName(destTagName)
	if err != nil {
		return fmt.Errorf("could not retrieve tag '%v': %v", destTagName, err)
	}
	if destTag == nil {
		return fmt.Errorf("no such tag '%v'.", destTagName)
	}

	for _, sourceTagName := range args[0 : len(args)-1] {
		if sourceTagName == destTagName {
			return fmt.Errorf("source and destination names are the same.")
		}

		sourceTag, err := store.TagByName(sourceTagName)
		if err != nil {
			return fmt.Errorf("could not retrieve tag '%v': %v", sourceTagName, err)
		}
		if sourceTag == nil {
			return fmt.Errorf("no such tag '%v'.", sourceTagName)
		}

		if command.verbose {
			log.Infof("finding files tagged '%v'.", sourceTagName)
		}

		explicitFileTags, err := store.ExplicitFileTagsByTagId(sourceTag.Id)
		if err != nil {
			return fmt.Errorf("could not retrieve explicit taggings for tag '%v': %v", sourceTagName, err)
		}

		implicitFileTags, err := store.ImplicitFileTagsByTagId(sourceTag.Id)
		if err != nil {
			return fmt.Errorf("could not retrieve implicit taggings for tag '%v': %v", sourceTagName, err)
		}

		if command.verbose {
			log.Infof("applying tag '%v' to these files.", destTagName)
		}

		for _, explicitFileTag := range explicitFileTags {
			_, err = store.AddExplicitFileTag(explicitFileTag.FileId, destTag.Id)
			if err != nil {
				return fmt.Errorf("could not add explicit tag '%v' to file #%v: %v", destTagName, explicitFileTag.FileId, err)
			}
		}

		for _, implicitFileTag := range implicitFileTags {
			_, err = store.AddImplicitFileTag(implicitFileTag.FileId, destTag.Id)
			if err != nil {
				return fmt.Errorf("could not add implicit tag '%v' to file #%v: %v", destTagName, implicitFileTag.FileId, err)
			}
		}

		if command.verbose {
			log.Infof("untagging files '%v'.", sourceTagName)
		}

		err = store.RemoveExplicitFileTagsByTagId(sourceTag.Id)
		if err != nil {
			return fmt.Errorf("could not remove all explicit taggings of tag '%v': %v", sourceTagName, err)
		}

		err = store.RemoveImplicitFileTagsByTagId(sourceTag.Id)
		if err != nil {
			return fmt.Errorf("could not remove all implicit taggings of tag '%v': %v", sourceTagName, err)
		}

		if command.verbose {
			log.Infof("deleting tag '%v'.", sourceTagName)
		}

		err = store.DeleteTag(sourceTag.Id)
		if err != nil {
			return fmt.Errorf("could not delete tag '%v': %v", sourceTagName, err)
		}
	}

	return nil
}
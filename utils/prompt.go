/*
Copyright © 2020 Alessandro Segala (@ItalyPaleAle)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package utils

import (
	"errors"

	"github.com/manifoldco/promptui"
)

func PromptMasterKey() (string, error) {
	prompt := promptui.Prompt{
		Validate: func(input string) error {
			if len(input) < 1 {
				return errors.New("Master key must not be empty")
			}
			return nil
		},
		Label: "Master key",
		Mask:  '*',
	}

	key, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return key, err
}

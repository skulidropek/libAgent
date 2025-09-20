//go:build tools

package tools

//go:generate go install github.com/skulidropek/GoSuggestMembersAnalyzer/cmd/smbgo

import (
	_ "github.com/skulidropek/GoSuggestMembersAnalyzer/smbgo"
)
